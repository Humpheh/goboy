// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

(() => {
	// Map web browser API and Node.js API to a single common API (preferring web standards over Node.js API).
	const isNodeJS = typeof process !== "undefined";
	if (isNodeJS) {
		global.require = require;
		global.fs = require("fs");

		const nodeCrypto = require("crypto");
		global.crypto = {
			getRandomValues(b) {
				nodeCrypto.randomFillSync(b);
			},
		};

		global.performance = {
			now() {
				const [sec, nsec] = process.hrtime();
				return sec * 1000 + nsec / 1000000;
			},
		};

		const util = require("util");
		global.TextEncoder = util.TextEncoder;
		global.TextDecoder = util.TextDecoder;
	} else {
		window.global = window;

		let outputBuf = "";
		global.fs = {
			constants: { O_WRONLY: -1, O_RDWR: -1, O_CREAT: -1, O_TRUNC: -1, O_APPEND: -1, O_EXCL: -1, O_NONBLOCK: -1, O_SYNC: -1 }, // unused
			writeSync(fd, buf) {
				outputBuf += decoder.decode(buf);
				const nl = outputBuf.lastIndexOf("\n");
				if (nl != -1) {
					console.log(outputBuf.substr(0, nl));
					outputBuf = outputBuf.substr(nl + 1);
				}
				return buf.length;
			},
			openSync(path, flags, mode) {
				const err = new Error("not implemented");
				err.code = "ENOSYS";
				throw err;
			},
		};
	}

	const encoder = new TextEncoder("utf-8");
	const decoder = new TextDecoder("utf-8");

	global.Go = class {
		constructor() {
			this.argv = ["js"];
			this.env = {};
			this.exit = (code) => {
				if (code !== 0) {
					console.warn("exit code:", code);
				}
			};
			this._callbackTimeouts = new Map();
			this._nextCallbackTimeoutID = 1;

			const mem = () => {
				// The buffer may change when requesting more memory.
				return new DataView(this._inst.exports.mem.buffer);
			}

			const setInt64 = (addr, v) => {
				mem().setUint32(addr + 0, v, true);
				mem().setUint32(addr + 4, Math.floor(v / 4294967296), true);
			}

			const getInt64 = (addr) => {
				const low = mem().getUint32(addr + 0, true);
				const high = mem().getInt32(addr + 4, true);
				return low + high * 4294967296;
			}

			const loadValue = (addr) => {
				const f = mem().getFloat64(addr, true);
				if (!isNaN(f)) {
					return f;
				}

				const id = mem().getUint32(addr, true);
				return this._values[id];
			}

			const storeValue = (addr, v) => {
				if (typeof v === "number") {
					if (isNaN(v)) {
						mem().setUint32(addr + 4, 0x7FF80000, true); // NaN
						mem().setUint32(addr, 0, true);
						return;
					}
					mem().setFloat64(addr, v, true);
					return;
				}

				mem().setUint32(addr + 4, 0x7FF80000, true); // NaN

				switch (v) {
					case undefined:
						mem().setUint32(addr, 1, true);
						return;
					case null:
						mem().setUint32(addr, 2, true);
						return;
					case true:
						mem().setUint32(addr, 3, true);
						return;
					case false:
						mem().setUint32(addr, 4, true);
						return;
				}

				if (typeof v === "string") {
					let ref = this._stringRefs.get(v);
					if (ref === undefined) {
						ref = this._values.length;
						this._values.push(v);
						this._stringRefs.set(v, ref);
					}
					mem().setUint32(addr, ref, true);
					return;
				}

				if (typeof v === "symbol") {
					let ref = this._symbolRefs.get(v);
					if (ref === undefined) {
						ref = this._values.length;
						this._values.push(v);
						this._symbolRefs.set(v, ref);
					}
					mem().setUint32(addr, ref, true);
					return;
				}

				let ref = v[this._refProp];
				if (ref === undefined) {
					ref = this._values.length;
					this._values.push(v);
					v[this._refProp] = ref;
				}
				mem().setUint32(addr, ref, true);
			}

			const loadSlice = (addr) => {
				const array = getInt64(addr + 0);
				const len = getInt64(addr + 8);
				return new Uint8Array(this._inst.exports.mem.buffer, array, len);
			}

			const loadSliceOfValues = (addr) => {
				const array = getInt64(addr + 0);
				const len = getInt64(addr + 8);
				const a = new Array(len);
				for (let i = 0; i < len; i++) {
					a[i] = loadValue(array + i * 8);
				}
				return a;
			}

			const loadString = (addr) => {
				const saddr = getInt64(addr + 0);
				const len = getInt64(addr + 8);
				return decoder.decode(new DataView(this._inst.exports.mem.buffer, saddr, len));
			}

			const timeOrigin = Date.now() - performance.now();
			this.importObject = {
				go: {
					// func wasmExit(code int32)
					"runtime.wasmExit": (sp) => {
						this.exited = true;
						this.exit(mem().getInt32(sp + 8, true));
					},

					// func wasmWrite(fd uintptr, p unsafe.Pointer, n int32)
					"runtime.wasmWrite": (sp) => {
						const fd = getInt64(sp + 8);
						const p = getInt64(sp + 16);
						const n = mem().getInt32(sp + 24, true);
						fs.writeSync(fd, new Uint8Array(this._inst.exports.mem.buffer, p, n));
					},

					// func nanotime() int64
					"runtime.nanotime": (sp) => {
						setInt64(sp + 8, (timeOrigin + performance.now()) * 1000000);
					},

					// func walltime() (sec int64, nsec int32)
					"runtime.walltime": (sp) => {
						const msec = (new Date).getTime();
						setInt64(sp + 8, msec / 1000);
						mem().setInt32(sp + 16, (msec % 1000) * 1000000, true);
					},

					// func scheduleCallback(delay int64) int32
					"runtime.scheduleCallback": (sp) => {
						const id = this._nextCallbackTimeoutID;
						this._nextCallbackTimeoutID++;
						this._callbackTimeouts.set(id, setTimeout(
							() => { this._resolveCallbackPromise(); },
							getInt64(sp + 8) + 1, // setTimeout has been seen to fire up to 1 millisecond early
						));
						mem().setInt32(sp + 16, id, true);
					},

					// func clearScheduledCallback(id int32)
					"runtime.clearScheduledCallback": (sp) => {
						const id = mem().getInt32(sp + 8, true);
						clearTimeout(this._callbackTimeouts.get(id));
						this._callbackTimeouts.delete(id);
					},

					// func getRandomData(r []byte)
					"runtime.getRandomData": (sp) => {
						crypto.getRandomValues(loadSlice(sp + 8));
					},

					// func stringVal(value string) ref
					"syscall/js.stringVal": (sp) => {
						storeValue(sp + 24, loadString(sp + 8));
					},

					// func valueGet(v ref, p string) ref
					"syscall/js.valueGet": (sp) => {
						storeValue(sp + 32, Reflect.get(loadValue(sp + 8), loadString(sp + 16)));
					},

					// func valueSet(v ref, p string, x ref)
					"syscall/js.valueSet": (sp) => {
						Reflect.set(loadValue(sp + 8), loadString(sp + 16), loadValue(sp + 32));
					},

					// func valueIndex(v ref, i int) ref
					"syscall/js.valueIndex": (sp) => {
						storeValue(sp + 24, Reflect.get(loadValue(sp + 8), getInt64(sp + 16)));
					},

					// valueSetIndex(v ref, i int, x ref)
					"syscall/js.valueSetIndex": (sp) => {
						Reflect.set(loadValue(sp + 8), getInt64(sp + 16), loadValue(sp + 24));
					},

					// func valueCall(v ref, m string, args []ref) (ref, bool)
					"syscall/js.valueCall": (sp) => {
						try {
							const v = loadValue(sp + 8);
							const m = Reflect.get(v, loadString(sp + 16));
							const args = loadSliceOfValues(sp + 32);
							storeValue(sp + 56, Reflect.apply(m, v, args));
							mem().setUint8(sp + 64, 1);
						} catch (err) {
							storeValue(sp + 56, err);
							mem().setUint8(sp + 64, 0);
						}
					},

					// func valueInvoke(v ref, args []ref) (ref, bool)
					"syscall/js.valueInvoke": (sp) => {
						try {
							const v = loadValue(sp + 8);
							const args = loadSliceOfValues(sp + 16);
							storeValue(sp + 40, Reflect.apply(v, undefined, args));
							mem().setUint8(sp + 48, 1);
						} catch (err) {
							storeValue(sp + 40, err);
							mem().setUint8(sp + 48, 0);
						}
					},

					// func valueNew(v ref, args []ref) (ref, bool)
					"syscall/js.valueNew": (sp) => {
						try {
							const v = loadValue(sp + 8);
							const args = loadSliceOfValues(sp + 16);
							storeValue(sp + 40, Reflect.construct(v, args));
							mem().setUint8(sp + 48, 1);
						} catch (err) {
							storeValue(sp + 40, err);
							mem().setUint8(sp + 48, 0);
						}
					},

					// func valueLength(v ref) int
					"syscall/js.valueLength": (sp) => {
						setInt64(sp + 16, parseInt(loadValue(sp + 8).length));
					},

					// valuePrepareString(v ref) (ref, int)
					"syscall/js.valuePrepareString": (sp) => {
						const str = encoder.encode(String(loadValue(sp + 8)));
						storeValue(sp + 16, str);
						setInt64(sp + 24, str.length);
					},

					// valueLoadString(v ref, b []byte)
					"syscall/js.valueLoadString": (sp) => {
						const str = loadValue(sp + 8);
						loadSlice(sp + 16).set(str);
					},

					// func valueInstanceOf(v ref, t ref) bool
					"syscall/js.valueInstanceOf": (sp) => {
						mem().setUint8(sp + 24, loadValue(sp + 8) instanceof loadValue(sp + 16));
					},

					"debug": (value) => {
						console.log(value);
					},
				}
			};
		}

		async run(instance) {
			this._inst = instance;
			this._values = [ // TODO: garbage collection
				NaN,
				undefined,
				null,
				true,
				false,
				global,
				this._inst.exports.mem,
				() => { // resolveCallbackPromise
					if (this.exited) {
						throw new Error("bad callback: Go program has already exited");
					}
					setTimeout(this._resolveCallbackPromise, 0); // make sure it is asynchronous
				},
			];
			this._stringRefs = new Map();
			this._symbolRefs = new Map();
			this._refProp = Symbol();
			this.exited = false;

			const mem = new DataView(this._inst.exports.mem.buffer)

			// Pass command line arguments and environment variables to WebAssembly by writing them to the linear memory.
			let offset = 4096;

			const strPtr = (str) => {
				let ptr = offset;
				new Uint8Array(mem.buffer, offset, str.length + 1).set(encoder.encode(str + "\0"));
				offset += str.length + (8 - (str.length % 8));
				return ptr;
			};

			const argc = this.argv.length;

			const argvPtrs = [];
			this.argv.forEach((arg) => {
				argvPtrs.push(strPtr(arg));
			});

			const keys = Object.keys(this.env).sort();
			argvPtrs.push(keys.length);
			keys.forEach((key) => {
				argvPtrs.push(strPtr(`${key}=${this.env[key]}`));
			});

			const argv = offset;
			argvPtrs.forEach((ptr) => {
				mem.setUint32(offset, ptr, true);
				mem.setUint32(offset + 4, 0, true);
				offset += 8;
			});

			while (true) {
				const callbackPromise = new Promise((resolve) => {
					this._resolveCallbackPromise = resolve;
				});
				this._inst.exports.run(argc, argv);
				if (this.exited) {
					break;
				}
				await callbackPromise;
			}
		}
	}

	if (isNodeJS) {
		if (process.argv.length < 3) {
			process.stderr.write("usage: go_js_wasm_exec [wasm binary] [arguments]\n");
			process.exit(1);
		}

		const go = new Go();
		go.argv = process.argv.slice(2);
		go.env = process.env;
		go.exit = process.exit;
		WebAssembly.instantiate(fs.readFileSync(process.argv[2]), go.importObject).then((result) => {
			process.on("exit", () => { // Node.js exits if no callback is pending
				if (!go.exited) {
					console.error("error: all goroutines asleep and no JavaScript callback pending - deadlock!");
					process.exit(1);
				}
			});
			return go.run(result.instance);
		}).catch((err) => {
			console.error(err);
			go.exited = true;
			process.exit(1);
		});
	}
})();

/*
var uuid = function() {
    function _p8(s) {
        var p = (Math.random().toString(16)+"000000000").substr(2,8);
        return s ? "-" + p.substr(0,4) + "-" + p.substr(4,4) : p ;
    }
    return _p8() + _p8(true) + _p8(true) + _p8();
};
*/

function PCMPlayer(ctx) {
  this.ctx = ctx;
  this.startTime = 0;
  this.numChannels = 2;
  this.sampleRate = 44100;
}

PCMPlayer.prototype.playNext = function (samplesLeft, samplesRight) {
  var audioBuffer = this.ctx.createBuffer(
    this.numChannels,
    samplesLeft.length,
    this.sampleRate
  );
  audioBuffer.getChannelData(0).set(samplesLeft);
  audioBuffer.getChannelData(1).set(samplesRight);

  if (this.startTime < this.ctx.currentTime) {
    this.startTime = this.ctx.currentTime;
  }

  var bufferSource = this.ctx.createBufferSource();
  bufferSource.buffer = audioBuffer;
  bufferSource.connect(this.ctx.destination);
  bufferSource.start(this.startTime);
  this.startTime += audioBuffer.duration;
};

// self.onload = function () {
(async function loadAndRunGoWasm() {
  const go = new Go();
  const response = await fetch("main.wasm");
  const buffer = await response.arrayBuffer();
  const result = await WebAssembly.instantiate(buffer, go.importObject);
  go.run(result.instance);
})();
// (() => {
//   if (!WebAssembly.instantiateStreaming) {
//     // polyfill
//     WebAssembly.instantiateStreaming = async (resp, importObject) => {
//       const source = await (await resp).arrayBuffer();
//       return await WebAssembly.instantiate(source, importObject);
//     };
//   }
//   const go = new Go();
//   let mod, inst;
//   WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject).then(
//     (result) => {
//       mod = result.module;
//       inst = result.instance;
//       run().then(
//         (result) => {
//           console.log("Ran WASM: ", result);
//         },
//         (failure) => {
//           console.log("Failed to run WASM: ", failure);
//         }
//       );
//     }
//   );
//   async function run() {
//     await go.run(inst);
//     inst = await WebAssembly.instantiate(mod, go.importObject); // reset instance
//   }
// })();
// };

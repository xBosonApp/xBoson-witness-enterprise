//
// 编译目录中的静态文件为 go 资源包
// 读取当前目录中的 build.json 作为编译配置
// 运行: 无参数执行该脚本 nodejs > v6
//
// 配置说明: 
//    遍历 wwwDir 目录中的文件, 将文件内容保存到 varName 变量中,
//    文件名是变量索引; 输出到 outDir/fileName 的 GO 源文件中,
//    包名为 packageName; 通常在包的其他源文件中定义 varName 变量,
//    变量类型是 map[string][]byte.
//    不支持深层目录中的文件.
//
const zb = require('zlib')
const fs = require('fs');
const pt = require('path');
const cf = require("./build.json");
const st = require("stream");

const go_code = `
//
// 这是程序生成的资源文件
//
import (
  "io/ioutil"
  "compress/gzip"
  "bytes"
  "log"
)

func _uz(input []byte) []byte {
  r, err := gzip.NewReader(bytes.NewBuffer(input))
  if err != nil {
    log.Println("Resource fail", err)
    return nil
  }
  a, err := ioutil.ReadAll(r)
  if err != nil {
    log.Println("Resource fail", err)
    return nil
  }
  return a
}
`

var fullpath = pt.join(__dirname, cf.outDir, cf.fileName);
var outfile = makeSource(fullpath, cf.varName);
outfile.setPackage(cf.packageName);
buildDir(pt.join(__dirname, cf.wwwDir), outfile);


function buildDir(dir, outfile) {
  var dirs = fs.readdirSync(dir);
  var i = -1;

  outfile.beginInit();
  _next();

  function _next() {
    if (++i < dirs.length) {
      var d = dirs[i];
      var file = pt.join(dir, d);
      outfile.localfile(file, d, _next);
      console.log('Local Resource:', file);
    } else {
      outfile.endInit();
      console.log("\nDone", outfile.fileName);
    }
  }
}


function makeSource(outFile, varName) {
  var file = fs.openSync(outFile, 'w');

  return {
    fileName   : outFile,
    setPackage : setPackage,
    localfile  : localfile,
    beginInit  : beginInit,
    endInit    : endInit,
  };

  function beginInit() {
    fs.writeSync(file, go_code);
    fs.writeSync(file, "\nfunc init() {");
  }

  function endInit() {
    fs.writeSync(file, "\n}");
  }

  function setPackage(pkName) {
    fs.writeSync(file, "package ");
    fs.writeSync(file, pkName);
    fs.writeSync(file, '\n');
  }

  function localfile(path, name, over) {
    fs.writeSync(file, ['\n\n', varName, '["', name, '"]=_uz([]byte{'].join(''));
    
    var wstream = fs.createWriteStream(null, {
      fd : file, 
      autoClose : false,
    });

    var r = fs.createReadStream(path);
    r.pipe(zb.createGzip())
      .pipe(byteArrEncode(end))
      .pipe(wstream);

    wstream.on('finish', end);
      
    function end() {
      fs.writeSync(file, '})');
      over();
    }
  }
}


//
// 把二进制写出为 go 语言字节数组
//
function byteArrEncode() {
  var enc = new st.Transform();
  enc._transform = function(chunk, encoding, callback) {
    for (var i=0; i<chunk.length; ++i) {
      var b = chunk[i];
      this.push(b.toString());
      this.push(',');
      if (i%20 == 0) this.push('\n');
    }
    callback();
  };
  return enc;
}
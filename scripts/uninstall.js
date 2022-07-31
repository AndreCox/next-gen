const os = require("os");
const fs = require("fs");

var platform = os.platform();

if (platform === "windows") {
  // remove the nextgen folder
  fs.rmSync(`${os.homedir()}/AppData/Local/nextgen`, { recursive: true });
} else if (platform === "linux") {
  // remove the nextgen folder
  fs.rmSync("/usr/bin/nextgen", { recursive: true });
} else if (platform === "darwin") {
  // remove the nextgen folder
  fs.rmSync("/usr/bin/nextgen", { recursive: true });
}

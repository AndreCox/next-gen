// get the os we are running on and the architecture
const os = require("os");
const https = require("https");
const fs = require("fs");
const dlgh = require("download-github-release");
const { spawnSync } = require("child_process");

var arch = os.arch();
var platform = os.platform();
var tag = "";
var fileExtension = ".tar.gz";

if (arch === "x64") {
  arch = "amd64";
}

if (platform === "win32") {
  platform = "windows";
  fileExtension = ".zip";
}

console.log("os: " + platform + " arch: " + arch);

let releasesUrl = "https://github.com/AndreCox/next-gen/releases/download/";

var options = {
  host: "api.github.com",
  path: "/repos/andrecox/next-gen/releases/latest",
  method: "GET",
  headers: { "user-agent": "node.js" },
};

// get the latest release tag
https.get(options, (res) => {
  res.setEncoding("utf8");
  let body = "";
  res.on("data", (data) => {
    body += data;
  });
  res.on("end", () => {
    body = JSON.parse(body);
    tag = body.tag_name;
    downloadGithubRelease("AndreCox", "next-gen", tag);
  });
});

const downloadGithubRelease = (user, repo, tag) => {
  function filterRelease(release) {
    // Filter out prereleases.
    return release.prerelease === false;
  }

  const filterAsset = (release) => {
    return (
      release.name === `nextgen-${tag}-${platform}-${arch}${fileExtension}`
    );
  };

  dlgh(user, repo, "./", filterRelease, filterAsset, true)
    .then(function () {
      console.log("Finished downloading");
      extract();
    })
    .catch(function (err) {
      console.error(err.message);
    });
};

const extract = () => {
  const decompress = require("decompress");
  if (platform === "windows") {
    decompress(`nextgen-${tag}-${platform}-${arch}.zip`, "./temp", {}).then(
      (files) => {
        console.log("Files extracted!");
        configurePaths();
        setPaths();
      }
    );
  } else {
    decompress(`nextgen-${tag}-${platform}-${arch}.tar.gz`, "./temp", {
      plugins: [decompressTargz()],
    }).then((files) => {
      console.log("Files extracted!");
      configurePaths();
      setPaths();
    });
  }
};

const configurePaths = () => {
  if (platform === "windows") {
    fs.mkdirSync(`${os.homedir()}/AppData/Local/nextgen`, { recursive: true });
    fs.copyFileSync(
      "./temp/nextgen.exe",
      `${os.homedir()}/AppData/Local/nextgen/nextgen.exe`
    );
  } else if (platform === "linux") {
    fs.mkdir("/usr/bin/nextgen");
    fs.copyFile("./temp/nextgen", "/usr/bin/nextgen");
  } else if (platform === "darwin") {
    fs.mkdir("/usr/bin/nextgen");
    fs.copyFile("./temp/nextgen", "/usr/bin/nextgen");
  }
};

const setPaths = () => {
  // export PATH=$PATH:/usr/local/bin
  console.log("setting paths");
  if (platform === "windows") {
    // create the nextgen folder
    var result = spawnSync("setx", [
      "PATH",
      os.homedir() + "/AppData/Local/nextgen",
    ]);
    console.log(result.stdout.toString());
  } else if (platform === "linux") {
    exec("export PATH=$PATH:/usr/bin/nextgen");
  } else if (platform === "darwin") {
    exec("export PATH=$PATH:/usr/bin/nextgen");
  }
  cleanUp();
};

const cleanUp = () => {
  // delete the temp folder
  fs.rmSync("./temp", { recursive: true });
  // delete the downloaded file
  if (platform === "windows") {
    fs.unlinkSync(`nextgen-${tag}-${platform}-${arch}.zip`);
  } else {
    fs.unlinkSync(`nextgen-${tag}-${platform}-${arch}.tar.gz`);
  }
};

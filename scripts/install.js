#!/usr/bin/env node

const fs = require("node:fs");
const path = require("node:path");
const { pipeline } = require("node:stream/promises");
const tar = require("tar");

const pkg = require("../package.json");

const rootDir = path.join(__dirname, "..");
const vendorDir = path.join(rootDir, "vendor");

function getTarget() {
  const platformMap = {
    linux: "linux",
    darwin: "darwin",
    win32: "windows",
  };

  const archMap = {
    x64: "amd64",
    arm64: "arm64",
  };

  const os = platformMap[process.platform];
  const arch = archMap[process.arch];
  if (!os || !arch) {
    throw new Error(`unsupported platform: ${process.platform}/${process.arch}`);
  }

  return { os, arch };
}

function binaryName() {
  return process.platform === "win32" ? "jtechforums.exe" : "jtechforums";
}

function releaseRepo() {
  return process.env.JTECHFORUMS_RELEASE_REPO || pkg.jtechforums.releaseRepo;
}

function releaseVersion() {
  const version = process.env.JTECHFORUMS_RELEASE_VERSION || pkg.version;
  return version.startsWith("v") ? version : `v${version}`;
}

function assetName() {
  const target = getTarget();
  return `jtechforums-${target.os}-${target.arch}.tar.gz`;
}

function assetURL() {
  return `https://github.com/${releaseRepo()}/releases/download/${releaseVersion()}/${assetName()}`;
}

async function download(url, destination) {
  const response = await fetch(url, {
    headers: { "user-agent": "jtechforums-installer" },
    redirect: "follow",
  });

  if (!response.ok || !response.body) {
    throw new Error(`download failed: ${response.status} ${response.statusText}`);
  }

  const file = fs.createWriteStream(destination);
  await pipeline(response.body, file);
}

async function install() {
  if (process.env.JTECHFORUMS_SKIP_DOWNLOAD === "1") {
    return;
  }

  const target = getTarget();
  const archive = path.join(vendorDir, assetName());
  const binaryPath = path.join(vendorDir, binaryName());

  fs.mkdirSync(vendorDir, { recursive: true });
  fs.rmSync(binaryPath, { force: true });
  fs.rmSync(archive, { force: true });

  try {
    await download(assetURL(), archive);
    await tar.x({
      cwd: vendorDir,
      file: archive,
      strict: true,
    });
    fs.chmodSync(binaryPath, 0o755);
    fs.rmSync(archive, { force: true });
    console.log(`installed jtechforums for ${target.os}/${target.arch}`);
  } catch (err) {
    fs.rmSync(archive, { force: true });
    throw err;
  }
}

async function main() {
  if (process.argv.includes("--print-target")) {
    const target = getTarget();
    console.log(JSON.stringify({ ...target, asset: assetName(), url: assetURL() }));
    return;
  }

  try {
    await install();
  } catch (err) {
    console.error(`jtechforums install failed: ${err.message}`);
    process.exit(1);
  }
}

main();

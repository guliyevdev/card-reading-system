#!/usr/bin/env node

const { spawnSync } = require("child_process");
const path = require("path");

const TARGET_NODE_VERSIONS = {
    node18: "18.5.0"
};

const PKG_SUPPORTED_RUNTIMES = Object.keys(TARGET_NODE_VERSIONS);

const [, , target, output] = process.argv;

if (!target || !output) {
    console.error("Usage: node scripts/build-pkg-target.js <pkg-target> <output>");
    process.exit(1);
}

const targetMatch = target.match(/^(node\d+)-(win|macos|linux)-(x64|arm64)$/);
if (!targetMatch) {
    console.error(`Unsupported pkg target: ${target}`);
    process.exit(1);
}

const [, runtime, targetPlatform, targetArch] = targetMatch;
const nodeVersion = TARGET_NODE_VERSIONS[runtime];

if (!nodeVersion) {
    console.error(
        [
            `Unsupported runtime ${runtime}.`,
            `This project currently builds with pkg-supported runtimes: ${PKG_SUPPORTED_RUNTIMES.join(", ")}.`,
            "pkg@5.8.1 does not provide a node24 base binary."
        ].join(" ")
    );
    process.exit(1);
}

const hostPlatformMap = {
    darwin: "macos",
    win32: "win",
    linux: "linux"
};

const hostPlatform = hostPlatformMap[process.platform];
if (!hostPlatform) {
    console.error(`Unsupported host platform: ${process.platform}`);
    process.exit(1);
}

if (hostPlatform !== targetPlatform || process.arch !== targetArch) {
    console.error(
        [
            `Cannot build ${target} on ${hostPlatform}-${process.arch}.`,
            "The @pokusew/pcsclite native module must be compiled on the same OS and CPU architecture as the final binary.",
            `Run this build on a ${targetPlatform}-${targetArch} machine instead.`
        ].join(" ")
    );
    process.exit(1);
}

const repoRoot = path.resolve(__dirname, "..");
const pcscliteDir = path.join(repoRoot, "node_modules", "@pokusew", "pcsclite");

function run(command, args, cwd) {
    const executable = process.platform === "win32" ? `${command}.cmd` : command;
    const result = spawnSync(executable, args, {
        cwd,
        stdio: "inherit"
    });

    if (result.status !== 0) {
        process.exit(result.status || 1);
    }
}

console.log(`Rebuilding @pokusew/pcsclite for ${target} (Node ${nodeVersion})`);
run("npx", [
    "node-gyp",
    "rebuild",
    `--target=${nodeVersion}`,
    `--arch=${targetArch}`,
    "--dist-url=https://nodejs.org/download/release"
], pcscliteDir);

console.log(`Packaging binary: ${output}`);
run("npx", [
    "pkg",
    ".",
    "--targets",
    target,
    "--output",
    output
], repoRoot);

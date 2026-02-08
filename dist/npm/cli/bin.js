#!/usr/bin/env node

"use strict";

const { spawnSync } = require("child_process");

const PLATFORMS = {
  darwin: {
    arm64: "@linterly/darwin-arm64/bin/linterly",
    x64: "@linterly/darwin-x64/bin/linterly",
  },
  linux: {
    arm64: "@linterly/linux-arm64/bin/linterly",
    x64: "@linterly/linux-x64/bin/linterly",
  },
  win32: {
    x64: "@linterly/win32-x64/bin/linterly.exe",
  },
};

function getBinaryPath() {
  const platform = PLATFORMS[process.platform];
  if (!platform) {
    throw new Error(
      `Unsupported platform: ${process.platform}. ` +
        `Linterly supports darwin, linux, and win32.`
    );
  }

  const binPath = platform[process.arch];
  if (!binPath) {
    throw new Error(
      `Unsupported architecture: ${process.arch} on ${process.platform}. ` +
        `Linterly supports ${Object.keys(platform).join(", ")} on ${process.platform}.`
    );
  }

  return binPath;
}

function main() {
  let binPath;
  try {
    binPath = require.resolve(getBinaryPath());
  } catch (e) {
    if (e.message && e.message.startsWith("Unsupported")) {
      console.error(`Error: ${e.message}`);
    } else {
      console.error(
        `Error: Could not find the Linterly binary for ${process.platform}-${process.arch}.\n` +
          `Please make sure the optional dependency is installed.\n` +
          `Try running: npm install @linterly/cli`
      );
    }
    process.exit(1);
  }

  const result = spawnSync(binPath, process.argv.slice(2), {
    stdio: "inherit",
  });

  if (result.error) {
    console.error(`Failed to execute Linterly: ${result.error.message}`);
    process.exit(1);
  }

  if (result.status !== null) {
    process.exit(result.status);
  }

  // シグナルで終了した場合: 128 + シグナル番号
  const SIGNALS = { SIGHUP: 1, SIGINT: 2, SIGTERM: 15, SIGKILL: 9 };
  process.exit(128 + (SIGNALS[result.signal] || 1));
}

main();

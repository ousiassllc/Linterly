#!/usr/bin/env node

const { platform, arch } = process;
const { spawnSync } = require("child_process");

const PLATFORMS = {
  win32: {
    x64: "@linterly/win32-x64/linterly.exe",
  },
  darwin: {
    x64: "@linterly/darwin-x64/linterly",
    arm64: "@linterly/darwin-arm64/linterly",
  },
  linux: {
    x64: "@linterly/linux-x64/linterly",
    arm64: "@linterly/linux-arm64/linterly",
  },
};

const binPath = PLATFORMS?.[platform]?.[arch];

if (!binPath) {
  console.error(
    `Unsupported platform: ${platform} ${arch}. ` +
      `The "@linterly/cli" package doesn't include a prebuilt binary for your platform.\n` +
      `Supported platforms: linux (x64, arm64), darwin (x64, arm64), win32 (x64).\n` +
      `You can build from source: https://github.com/ousiassllc/linterly`
  );
  process.exitCode = 1;
} else {
  let resolvedPath;
  try {
    resolvedPath = require.resolve(binPath);
  } catch {
    console.error(
      `The platform-specific package for "${platform}-${arch}" is not installed.\n` +
        `Expected package path: ${binPath}\n\n` +
        `If you installed with --no-optional, reinstall without that flag:\n` +
        `  npm install @linterly/cli\n\n` +
        `If the problem persists, try removing node_modules and reinstalling.`
    );
    process.exitCode = 1;
  }

  if (resolvedPath) {
    const result = spawnSync(resolvedPath, process.argv.slice(2), {
      shell: false,
      stdio: "inherit",
    });

    if (result.error) {
      throw result.error;
    }

    if (result.signal) {
      process.exitCode = 1;
    } else {
      process.exitCode = result.status;
    }
  }
}

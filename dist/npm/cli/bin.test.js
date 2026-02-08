"use strict";

const { describe, it, afterEach } = require("node:test");
const assert = require("node:assert/strict");
const { PLATFORMS, getBinaryPath } = require("./bin.js");

describe("PLATFORMS", () => {
  it("darwin-arm64 のパスが正しい", () => {
    assert.equal(
      PLATFORMS.darwin.arm64,
      "@linterly/darwin-arm64/bin/linterly"
    );
  });

  it("darwin-x64 のパスが正しい", () => {
    assert.equal(PLATFORMS.darwin.x64, "@linterly/darwin-x64/bin/linterly");
  });

  it("linux-arm64 のパスが正しい", () => {
    assert.equal(PLATFORMS.linux.arm64, "@linterly/linux-arm64/bin/linterly");
  });

  it("linux-x64 のパスが正しい", () => {
    assert.equal(PLATFORMS.linux.x64, "@linterly/linux-x64/bin/linterly");
  });

  it("win32-x64 のパスが正しい", () => {
    assert.equal(PLATFORMS.win32.x64, "@linterly/win32-x64/bin/linterly.exe");
  });

  it("Windows バイナリは .exe 拡張子を持つ", () => {
    for (const path of Object.values(PLATFORMS.win32)) {
      assert.ok(path.endsWith(".exe"), `${path} should end with .exe`);
    }
  });

  it("非 Windows バイナリは .exe 拡張子を持たない", () => {
    for (const [os, archs] of Object.entries(PLATFORMS)) {
      if (os === "win32") continue;
      for (const path of Object.values(archs)) {
        assert.ok(!path.endsWith(".exe"), `${path} should not end with .exe`);
      }
    }
  });
});

describe("getBinaryPath", () => {
  const originalPlatform = Object.getOwnPropertyDescriptor(
    process,
    "platform"
  );
  const originalArch = Object.getOwnPropertyDescriptor(process, "arch");

  afterEach(() => {
    if (originalPlatform) {
      Object.defineProperty(process, "platform", originalPlatform);
    }
    if (originalArch) {
      Object.defineProperty(process, "arch", originalArch);
    }
  });

  function mockPlatform(platform, arch) {
    Object.defineProperty(process, "platform", { value: platform });
    Object.defineProperty(process, "arch", { value: arch });
  }

  it("linux-x64 で正しいパスを返す", () => {
    mockPlatform("linux", "x64");
    assert.equal(getBinaryPath(), "@linterly/linux-x64/bin/linterly");
  });

  it("darwin-arm64 で正しいパスを返す", () => {
    mockPlatform("darwin", "arm64");
    assert.equal(getBinaryPath(), "@linterly/darwin-arm64/bin/linterly");
  });

  it("win32-x64 で正しいパスを返す", () => {
    mockPlatform("win32", "x64");
    assert.equal(getBinaryPath(), "@linterly/win32-x64/bin/linterly.exe");
  });

  it("未対応プラットフォームでエラーを投げる", () => {
    mockPlatform("freebsd", "x64");
    assert.throws(() => getBinaryPath(), {
      message: /Unsupported platform: freebsd/,
    });
  });

  it("未対応アーキテクチャでエラーを投げる", () => {
    mockPlatform("linux", "ia32");
    assert.throws(() => getBinaryPath(), {
      message: /Unsupported architecture: ia32 on linux/,
    });
  });

  it("win32-arm64 で未対応アーキテクチャエラーを投げる", () => {
    mockPlatform("win32", "arm64");
    assert.throws(() => getBinaryPath(), {
      message: /Unsupported architecture: arm64 on win32/,
    });
  });

  it("未対応プラットフォームのエラーメッセージに対応 OS を含む", () => {
    mockPlatform("freebsd", "x64");
    assert.throws(() => getBinaryPath(), {
      message: /darwin, linux, and win32/,
    });
  });

  it("未対応アーキテクチャのエラーメッセージに対応アーキテクチャを含む", () => {
    mockPlatform("darwin", "ia32");
    assert.throws(() => getBinaryPath(), {
      message: /arm64, x64/,
    });
  });
});

class Specwatch < Formula
  desc "Spec-driven architectural drift detector for the terminal"
  homepage "https://github.com/RajeshShrirao/Specwatch"
  version "0.0.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/RajeshShrirao/Specwatch/releases/download/v0.0.0/specwatch_macOS_arm64.tar.gz"
      sha256 "REPLACE_WITH_MACOS_ARM64_SHA256"
    else
      url "https://github.com/RajeshShrirao/Specwatch/releases/download/v0.0.0/specwatch_macOS_x86_64.tar.gz"
      sha256 "REPLACE_WITH_MACOS_AMD64_SHA256"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/RajeshShrirao/Specwatch/releases/download/v0.0.0/specwatch_Linux_arm64.tar.gz"
      sha256 "REPLACE_WITH_LINUX_ARM64_SHA256"
    else
      url "https://github.com/RajeshShrirao/Specwatch/releases/download/v0.0.0/specwatch_Linux_x86_64.tar.gz"
      sha256 "REPLACE_WITH_LINUX_AMD64_SHA256"
    end
  end

  def install
    bin.install "specwatch"
  end

  test do
    assert_match "specwatch version", shell_output("#{bin}/specwatch version")
  end
end

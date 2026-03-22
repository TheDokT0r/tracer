class Tracer < Formula
  desc "TUI for browsing, inspecting, resuming, and deleting Claude Code sessions"
  homepage "https://github.com/TheDokT0r/tracer"
  version "0.12.1"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/TheDokT0r/tracer/releases/download/v0.12.1/tracer-darwin-arm64.tar.gz"
      sha256 "660252af052a51fdfcc4854165e84cd0ff2ec95237ea1f88c32f5e228e14bf2d"
    end
    on_intel do
      url "https://github.com/TheDokT0r/tracer/releases/download/v0.12.1/tracer-darwin-amd64.tar.gz"
      sha256 "97482ffefe5ae5af02568d55368aa426648e210e94bcb294cfee6020644f4ed9"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/TheDokT0r/tracer/releases/download/v0.12.1/tracer-linux-arm64.tar.gz"
      sha256 "017cbdef6243b59c2eb8c56908cd9dc0a76c8583b72a006af63ff03a4bc15e88"
    end
    on_intel do
      url "https://github.com/TheDokT0r/tracer/releases/download/v0.12.1/tracer-linux-amd64.tar.gz"
      sha256 "311f5d037056a1bca3a420486e7872fcbd69398892151b6dd22dc384bd01388e"
    end
  end

  def install
    bin.install "tracer"
    man1.install "tracer.1"
  end

  test do
    assert_match "tracer", shell_output("#{bin}/tracer --version")
  end
end

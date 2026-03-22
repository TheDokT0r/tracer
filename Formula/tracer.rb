class Tracer < Formula
  desc "TUI for browsing, inspecting, resuming, and deleting Claude Code sessions"
  homepage "https://github.com/TheDokT0r/tracer"
  version "0.12.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/TheDokT0r/tracer/releases/download/v0.12.0/tracer-darwin-arm64.tar.gz"
      sha256 "394fdf8a4b461425e2353ed08ae321940f547c2cbfaf6ed71d7ed2420c6fdb0f"
    end
    on_intel do
      url "https://github.com/TheDokT0r/tracer/releases/download/v0.12.0/tracer-darwin-amd64.tar.gz"
      sha256 "1e45c5054956db7669ffc14780fe83cb1196fe210363a2ab895a544dbe2f435c"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/TheDokT0r/tracer/releases/download/v0.12.0/tracer-linux-arm64.tar.gz"
      sha256 "61645d12f87a3f9500c361df25eea87b39be146065fbd8dd8df505af1cb72ee3"
    end
    on_intel do
      url "https://github.com/TheDokT0r/tracer/releases/download/v0.12.0/tracer-linux-amd64.tar.gz"
      sha256 "e8abb284316887000923331ceafe57e152788c2403b97b1e4f57cf085206cbb7"
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

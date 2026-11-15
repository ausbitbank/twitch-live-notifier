class TwitchLiveNotifier < Formula
  desc "Twitch live alerts for labelled Discord webhooks"
  homepage "https://github.com/ausbitbank/twitch-live-notifier"
  url "https://github.com/ausbitbank/twitch-live-notifier/archive/refs/tags/v__VERSION__.tar.gz"
  sha256 "__SHA256__"
  license "MIT"
  head "https://github.com/ausbitbank/twitch-live-notifier.git", branch: "main"

  depends_on "go" => :build

  def install
    ldflags = "-s -w -X main.Version=#{version}"
    system "go", "build", "-trimpath", "-buildvcs=false",
           "-ldflags", ldflags,
           "-o", bin/"TwitchLiveNotifier", "."
  end

  service do
    run [opt_bin/"TwitchLiveNotifier", "--serve"]
    keep_alive true
    log_path var/"log/twitch-live-notifier.log"
    error_log_path var/"log/twitch-live-notifier.log"
  end

  def caveats
    <<~EOS
      Start the dashboard service with:
        brew services start twitch-live-notifier

      Then open:
        http://127.0.0.1:8787/

      Settings and secrets live in:
        ~/Library/Application Support/TwitchLiveNotifier
    EOS
  end

  test do
    assert_match "Twitch Live to Discord", shell_output("#{bin}/TwitchLiveNotifier --version")
  end
end

# 📱 BenzCloud Client

> **Notice:** BenzCloud Client is in **active development / pre-release status** (`v1.0`).

BenzCloud Client is the cross-platform client companion for the BenzCloud enterprise suite. It connects devices to the secure Slack Nebula Mesh-VPN overlay and built-in Custom DNS server, offering single-click access to all enterprise services (Drive, Mail, Chat, and custom websites).

## ✨ Features

- **Instant Zero-Config Pairing:** Connects initially to the BenzCloud Server LAN IP (e.g. `http://192.168.0.5:8080`), automatically downloads signed Nebula certificates and configuration, and transitions all traffic into the encrypted P2P mesh network.
- **Service Launcher Dashboard:** Fast access to:
  - 📁 **BenzCloud Drive** (`http://drive.<domain>`)
  - 📧 **BenzCloud Mail** (`http://mail.<domain>`)
  - 💬 **BenzCloud Chat** (`http://chat.<domain>`)
  - 🌐 **Websites & Portals** (`http://<domain>`)
- **Immutable Privileges:** Fully integrates with the server's guarantee that Mesh-VPN and DNS connectivity are permanently active for all accounts.
- **Multi-Platform:** Builds for Linux (WebKitGTK desktop shell), Windows (App-Mode executable), and native Android APK.

---

## 📦 Installation & Usage

### Linux (x86_64)
```bash
# Launch client desktop GUI
./benzcloud-client

# Or pair via command line
./benzcloud-client -server http://192.168.0.5:8080 -user admin -pass YourPassword
```

### Windows (x86_64)
Launch `benzcloud-client.exe` to connect and browse your local enterprise cloud.

### Android
Install `benzcloud-client-v1.0.apk` for mobile access to your files, email, and chats.

---

## 👥 Authors & Credits
- **Jeremy Benz** ([@benzjeremy](https://github.com/benzjeremy)) – Lead Engineer & Project Creator
- Pair-programmed with AI Assistant (Google Antigravity)
- © 2026 Jeremy Benz

## 📄 License & Third-Party Notices
- **Main Project:** Released under the [GNU General Public License v3.0 (GPL-3.0)](LICENSE).
- **Slack Nebula:** Mesh-VPN networking is powered by Slack Nebula, licensed under the [MIT License](https://github.com/slackhq/nebula/blob/master/LICENSE).

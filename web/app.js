// BenzCloud Client App Controller
let currentLang = localStorage.getItem("benzcloud_client_lang") || "de";
let clientProfile = null;

const i18n = {
  de: {
    pair_badge: "🔗 Client-Ersteinrichtung",
    pair_title: "Mit BenzCloud-Server verbinden",
    pair_desc: "Gib die lokale IP-Adresse deines BenzCloud-Servers und deine Anmeldedaten ein. Das Mesh-VPN (Nebula) und der DNS-Server konfigurieren sich automatisch.",
    lbl_server_url: "Server LAN-Adresse:",
    lbl_user: "Benutzername:",
    lbl_pass: "Passwort:",
    btn_pair: "🚀 Gerät sicher koppeln & verbinden",
    status_connected: "Verbunden mit BenzCloud",
    services_heading: "Deine Enterprise-Dienste (Direktzugriff)",
    card_drive_title: "BenzCloud Drive",
    card_drive_desc: "Dateien hochladen, teilen und synchronisieren.",
    card_mail_title: "BenzCloud Mail",
    card_mail_desc: "Internes Webmail & Thunderbird-kompatibles Postfach.",
    card_chat_title: "BenzCloud Chat",
    card_chat_desc: "Echtzeit-Teamkommunikation & Direktnachrichten.",
    card_web_title: "Webseiten & Portal",
    card_web_desc: "Gehostete Firmen- und Team-Webseiten aufrufen.",
    btn_unpair: "🔌 Entkoppeln / Gerät trennen",
    pairing_in_progress: "⚙️ Kopplung läuft...",
    confirm_unpair: "Möchtest du dieses Gerät wirklich vom BenzCloud-Server trennen?"
  },
  en: {
    pair_badge: "🔗 Initial Client Setup",
    pair_title: "Connect to BenzCloud Server",
    pair_desc: "Enter your local BenzCloud server IP and your credentials. Mesh-VPN (Nebula) and DNS server will configure automatically.",
    lbl_server_url: "Server LAN Address:",
    lbl_user: "Username:",
    lbl_pass: "Password:",
    btn_pair: "🚀 Securely Pair & Connect Device",
    status_connected: "Connected to BenzCloud",
    services_heading: "Your Enterprise Services (Direct Access)",
    card_drive_title: "BenzCloud Drive",
    card_drive_desc: "Upload, share, and synchronize cloud files.",
    card_mail_title: "BenzCloud Mail",
    card_mail_desc: "Internal webmail & Thunderbird-compatible mailbox.",
    card_chat_title: "BenzCloud Chat",
    card_chat_desc: "Real-time team communication & direct messages.",
    card_web_title: "Websites & Portal",
    card_web_desc: "Browse hosted team and company websites.",
    btn_unpair: "🔌 Disconnect / Unpair Device",
    pairing_in_progress: "⚙️ Pairing in progress...",
    confirm_unpair: "Are you sure you want to disconnect this device from BenzCloud?"
  }
};

document.addEventListener("DOMContentLoaded", () => {
  applyLanguage(currentLang);
  fetchClientProfile();
});

function setLanguage(lang) {
  currentLang = lang;
  localStorage.setItem("benzcloud_client_lang", lang);
  applyLanguage(lang);
}

function applyLanguage(lang) {
  document.documentElement.lang = lang;
  document.querySelectorAll("[data-i18n]").forEach(el => {
    const key = el.getAttribute("data-i18n");
    if (i18n[lang] && i18n[lang][key]) {
      el.textContent = i18n[lang][key];
    }
  });
  document.getElementById("langDE").classList.toggle("active", lang === "de");
  document.getElementById("langEN").classList.toggle("active", lang === "en");
}

async function fetchClientProfile() {
  try {
    const res = await fetch("/api/profile");
    if (!res.ok) {
      showPairingView();
      return;
    }
    const data = await res.json();
    if (!data.paired) {
      showPairingView();
      return;
    }
    clientProfile = data;
    showConnectedView(data);
  } catch (err) {
    showPairingView();
  }
}

function showPairingView() {
  document.getElementById("pairingView").style.display = "block";
  document.getElementById("connectedView").style.display = "none";
}

function showConnectedView(data) {
  document.getElementById("pairingView").style.display = "none";
  document.getElementById("connectedView").style.display = "block";

  document.getElementById("dispDomain").textContent = data.base_domain || "intern";
  document.getElementById("dispOverlayIP").textContent = data.overlay_ip || "10.42.0.2";
  document.getElementById("dispLighthouse").textContent = data.server_vpn_ip || "10.42.0.1";

  const domain = data.base_domain;
  document.getElementById("subDrive").textContent = `http://drive.${domain}`;
  document.getElementById("cardDrive").href = `http://drive.${domain}`;

  document.getElementById("subMail").textContent = `http://mail.${domain}`;
  document.getElementById("cardMail").href = `http://mail.${domain}`;

  document.getElementById("subChat").textContent = `http://chat.${domain}`;
  document.getElementById("cardChat").href = `http://chat.${domain}`;

  document.getElementById("subWeb").textContent = `http://${domain}`;
  document.getElementById("cardWeb").href = `http://${domain}`;
}

async function submitPairing(e) {
  e.preventDefault();
  const serverUrl = document.getElementById("serverUrl").value.trim();
  const username = document.getElementById("pairUser").value.trim();
  const password = document.getElementById("pairPass").value;

  const btn = document.getElementById("btnPair");
  btn.disabled = true;
  btn.textContent = i18n[currentLang].pairing_in_progress;

  try {
    const res = await fetch("/api/pair", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ server_url: serverUrl, username, password })
    });
    const data = await res.json();
    if (!res.ok) {
      alert("Kopplung fehlgeschlagen: " + (data.error || "Serverfehler"));
      btn.disabled = false;
      btn.textContent = i18n[currentLang].btn_pair;
      return;
    }
    fetchClientProfile();
  } catch (err) {
    alert("Netzwerkfehler: " + err);
    btn.disabled = false;
    btn.textContent = i18n[currentLang].btn_pair;
  }
}

async function unpairClient() {
  if (!confirm(i18n[currentLang].confirm_unpair)) return;
  try {
    await fetch("/api/unpair", { method: "POST" });
    fetchClientProfile();
  } catch (err) {
    alert("Fehler beim Entkoppeln");
  }
}

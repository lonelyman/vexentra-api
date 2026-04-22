# 🤖 Local LLM Coding Assistant — Mac Mini M1 16GB

## สรุป Setup

| | รายละเอียด |
|---|---|
| **Hardware** | Mac Mini M1 — 8-Core, 16GB Unified Memory |
| **Model** | Qwen2.5-Coder-14B-Instruct |
| **Quantization** | Q5_K_M |
| **ขนาดไฟล์** | ~10.2 GB |
| **Context Window** | 32,768 tokens |
| **ความเร็ว** | ~10–12 tok/s |
| **การใช้งาน** | Dedicated LLM Server — ไม่รันอะไรเพิ่ม |

---

## ทำไมถึงเลือกตัวนี้

- **14B** — quality ดีกว่า 7B พอสมควร เหมาะงาน Go / PHP
- **Q5_K_M** — เครื่อง dedicated รัน AI อย่างเดียว RAM เหลือพอ ไม่ swap
- **32K context** — รับโค้ดไฟล์ใหญ่ได้ ไม่ตัดกลางคัน
- **Offline 100%** — โค้ดไม่ออกไปไหน ปลอดภัย

---

## ขั้นตอนติดตั้ง

### 1. ติดตั้ง Ollama

```bash
brew install ollama
```

### 2. Pull Model

```bash
ollama pull qwen2.5-coder:14b-instruct-q5_K_M
```

> ⏳ ใช้เวลาโหลด ~10.2 GB ขึ้นอยู่กับความเร็ว internet

### 3. สร้าง Modelfile พร้อม Context Window 32K

```bash
cat > ~/Modelfile << EOF
FROM qwen2.5-coder:14b-instruct-q5_K_M
PARAMETER num_ctx 32768
EOF

ollama create qwen-coder-14b -f ~/Modelfile
```

### 4. เปิด Ollama ให้รับ Connection จากทั้ง Network

```bash
OLLAMA_HOST=0.0.0.0 ollama serve
```

> 📌 เปิดทิ้งไว้ตลอด เครื่องนี้เสียบไฟอยู่แล้ว

---

## ใช้งานจาก MacBook (หรือเครื่องอื่นในออฟฟิศ)

### เช็ค IP ของ Mac Mini ก่อน

```bash
# รันบน Mac Mini
ipconfig getifaddr en0
# ได้ IP เช่น 192.168.1.100
```

### ตั้ง Continue.dev ให้ชี้มาที่ Mac Mini

```json
// ~/.continue/config.json
{
  "models": [
    {
      "title": "Qwen2.5-Coder 14B (Mac Mini)",
      "provider": "ollama",
      "model": "qwen-coder-14b",
      "apiBase": "http://192.168.1.100:11434"
    }
  ],
  "tabAutocompleteModel": {
    "title": "Qwen2.5-Coder 14B",
    "provider": "ollama",
    "model": "qwen-coder-14b",
    "apiBase": "http://192.168.1.100:11434"
  }
}
```

> เปลี่ยน `192.168.1.100` เป็น IP จริงของ Mac Mini

---

## ทดสอบว่า Server ทำงานอยู่

```bash
# รันจากเครื่องไหนก็ได้
curl http://192.168.1.100:11434/api/generate \
  -d '{
    "model": "qwen-coder-14b",
    "prompt": "write a Go function to reverse a string",
    "stream": false
  }'
```

---

## ข้อแนะนำการใช้งาน

- ✅ ต่อ **LAN** ดีกว่า WiFi — latency ต่ำกว่า response ไว
- ✅ **ไม่ต้องเปิด app อื่น** บน Mac Mini — dedicated ให้ AI อย่างเดียว
- ✅ ถ้า response ช้าผิดปกติ เช็ค RAM pressure ด้วย Activity Monitor
- ⚠️ ถ้าเพิ่ม context เกิน 32K จะกิน RAM มากขึ้น อาจ swap

---

## RAM Usage โดยประมาณ

| ส่วน | RAM |
|---|---|
| macOS | ~3–4 GB |
| Model Q5_K_M | ~10.2 GB |
| Context 32K | ~2–3 GB |
| **รวม** | **~15–16 GB** ✅ |

---

## ถ้าอยากลอง Upgrade ในอนาคต

| Hardware | Model ที่รันได้ | คุณภาพ |
|---|---|---|
| M1/M2 16GB (ปัจจุบัน) | 14B Q5_K_M | ดี |
| M2 Pro 32GB | 32B Q4_K_M | ดีมาก |
| M3 Max 64GB | 32B Q8 หรือ 70B Q4 | ยอดเยี่ยม |

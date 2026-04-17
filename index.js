const express = require("express");
const cors = require("cors");
const { NFC } = require("nfc-pcsc");

const app = express();
app.use(cors());

let lastUID = null;
let lastATR = null;

const nfc = new NFC();

async function readUid(reader) {
    const cmd = Buffer.from("FFCA000000", "hex");
    const response = await reader.transmit(cmd, 12);

    if (!response || response.length < 2) {
        throw new Error(`Invalid UID response length: ${response ? response.length : 0}`);
    }

    const statusCode = response.slice(-2).readUInt16BE(0);
    if (statusCode !== 0x9000) {
        throw new Error(`Could not get card UID. Status=0x${statusCode.toString(16).toUpperCase()}`);
    }

    return response.slice(0, -2).toString("hex").toUpperCase();
}

nfc.on("reader", reader => {
    console.log(`Reader detected: ${reader.name}`);
    reader.autoProcessing = false;

    reader.on("card", async card => {
        try {
            lastATR = card.atr
                ? card.atr.toString("hex").toUpperCase()
                : null;

            lastUID = await readUid(reader);

            console.log(`Card detected. UID=${lastUID}, ATR=${lastATR}, standard=${card.standard || "unknown"}`);
        } catch (err) {
            lastUID = null;
            console.error("Card processing error:", err.message || err);
        }
    });

    reader.on("card.off", () => {
        console.log("Card removed");

        lastUID = null;
        lastATR = null;
    });

    reader.on("error", err => {
        console.error(`Reader error (${reader.name}):`, err.message || err);
    });

    reader.on("end", () => {
        console.log(`Reader removed: ${reader.name}`);

        lastUID = null;
        lastATR = null;
    });
});

nfc.on("error", err => {
    console.error("NFC error:", err.message || err);
});

// API endpoint (same as your original)
app.get("/card", (req, res) => {
    res.json({
        uid: lastUID,
        atr: lastATR
    });
});

app.listen(4121, () => {
    console.log("Local smartcard server running on http://localhost:4121");
});

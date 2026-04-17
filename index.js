const express = require("express");
const cors = require("cors");
const { NFC } = require("nfc-pcsc");

const app = express();
app.use(cors());

let lastUID = null;
let lastATR = null;

const nfc = new NFC();

nfc.on("reader", reader => {
    console.log(`Reader detected: ${reader.name}`);

    reader.on("card", card => {
        try {
            // UID comes directly from nfc-pcsc
            lastUID = card.uid ? card.uid.toUpperCase() : null;

            // ATR (if available)
            lastATR = card.atr
                ? card.atr.toString("hex").toUpperCase()
                : null;

            console.log(`Card detected. UID=${lastUID}, ATR=${lastATR}`);
        } catch (err) {
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
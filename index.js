const express = require("express");
const cors = require("cors");
const pcsclite = require("pcsclite");

const app = express();
app.use(cors());

let lastUID = null;  // UIDI burada saxlayırıq
let lastATR = null;

const pcsc = pcsclite();

pcsc.on("reader", reader => {

    reader.on("status", status => {
        const changes = reader.state ^ status.state;

        // Kart daxil oldu
        if (changes & reader.SCARD_STATE_PRESENT && status.state & reader.SCARD_STATE_PRESENT) {
            reader.connect({ share_mode: reader.SCARD_SHARE_SHARED }, (err, protocol) => {
                if (err) return console.error("Connect error:", err);

                // UID oxuma APDU
                const cmd = Buffer.from("FFCA000000", "hex");

                reader.transmit(cmd, 40, protocol, (err, data) => {
                    if (!err) {
                        lastUID = data.toString("hex").toUpperCase();
                    }
                });

                // ATR oxumaq üçün
                lastATR = status.atr ? status.atr.toString("hex").toUpperCase() : null;
            });
        }

        // Kart çıxarıldı
        if (changes & reader.SCARD_STATE_EMPTY && status.state & reader.SCARD_STATE_EMPTY) {
            lastUID = null;
        }
    });

});

// Frontend üçün API
app.get("/card", (req, res) => {
    res.json({
        uid: lastUID,
        atr: lastATR
    });
});

app.listen(4121, () => {
    console.log("Local smartcard server running on http://localhost:4121");
});
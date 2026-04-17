const express = require("express");
const cors = require("cors");
const pcsclite = require("pcsclite");

const app = express();
app.use(cors());

let lastUID = null;
let lastATR = null;

const pcsc = pcsclite();

pcsc.on("reader", reader => {
    let activeProtocol = null;

    console.log(`Reader detected: ${reader.name}`);

    reader.on("error", err => {
        console.error(`Reader error (${reader.name}):`, err.message || err);
    });

    reader.on("end", () => {
        console.log(`Reader removed: ${reader.name}`);
        lastUID = null;
        lastATR = null;
        activeProtocol = null;
    });

    reader.on("status", status => {
        const changes = reader.state ^ status.state;
        reader.state = status.state;

        if (changes & reader.SCARD_STATE_PRESENT && status.state & reader.SCARD_STATE_PRESENT) {
            lastATR = status.atr ? status.atr.toString("hex").toUpperCase() : null;

            reader.connect({ share_mode: reader.SCARD_SHARE_SHARED }, (err, protocol) => {
                if (err) {
                    console.error("Connect error:", err.message || err);
                    return;
                }

                if (Number.isInteger(protocol)) {
                    activeProtocol = protocol;
                }

                if (!Number.isInteger(activeProtocol)) {
                    console.error("Transmit skipped: protocol is unavailable");
                    return;
                }

                const cmd = Buffer.from("FFCA000000", "hex");

                reader.transmit(cmd, 40, activeProtocol, (err, data) => {
                    if (err) {
                        console.error("Transmit error:", err.message || err);
                        return;
                    }

                    lastUID = data.toString("hex").toUpperCase();
                    console.log(`Card detected. UID=${lastUID}, ATR=${lastATR}`);
                });
            });
        }

        if (changes & reader.SCARD_STATE_EMPTY && status.state & reader.SCARD_STATE_EMPTY) {
            lastUID = null;
            lastATR = null;
            activeProtocol = null;

            reader.disconnect(reader.SCARD_LEAVE_CARD, err => {
                if (err) {
                    console.error("Disconnect error:", err.message || err);
                    return;
                }

                console.log("Card removed");
            });
        }
    });
});

pcsc.on("error", err => {
    console.error("PCSC error:", err.message || err);
});

app.get("/card", (req, res) => {
    res.json({
        uid: lastUID,
        atr: lastATR
    });
});

app.listen(4121, () => {
    console.log("Local smartcard server running on http://localhost:4121");
});

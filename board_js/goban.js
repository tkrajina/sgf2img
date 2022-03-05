var _a;
var _b;
var MIDDOT = ".";
var BLACK_CIRCLE = "b";
var WHITE_CIRCLE = "w";
var SGFTag;
(function (SGFTag) {
    SGFTag["AddBlack"] = "AB";
    SGFTag["AddWhite"] = "AW";
    SGFTag["Annotations"] = "AN";
    SGFTag["Application"] = "AP";
    SGFTag["BlackMove"] = "B";
    SGFTag["BlackRank"] = "BR";
    SGFTag["BlackTeam"] = "BT";
    SGFTag["Comment"] = "C";
    SGFTag["Copyright"] = "CP";
    SGFTag["Date"] = "DT";
    SGFTag["Event"] = "EV";
    SGFTag["FileFormat"] = "FF";
    SGFTag["Game"] = "GM";
    SGFTag["GameName"] = "GN";
    SGFTag["Handicap"] = "HA";
    SGFTag["Komi"] = "KM";
    SGFTag["Opening"] = "ON";
    SGFTag["Overtime"] = "OT";
    SGFTag["BlackName"] = "PB";
    SGFTag["Place"] = "PC";
    SGFTag["Player"] = "PL";
    SGFTag["WhiteName"] = "PW";
    SGFTag["Result"] = "RE";
    SGFTag["Round"] = "RO";
    SGFTag["Rules"] = "RU";
    SGFTag["Source"] = "SO";
    SGFTag["Size"] = "SZ";
    SGFTag["Timelimit"] = "TM";
    SGFTag["User"] = "US";
    SGFTag["WhiteMove"] = "W";
    SGFTag["WhiteRank"] = "WR";
    SGFTag["WhiteTeam"] = "WT";
    // Additional
    SGFTag["Triangle"] = "TR";
    SGFTag["Square"] = "SQ";
    SGFTag["Circle"] = "CR";
    SGFTag["X"] = "MA";
    SGFTag["Label"] = "LB";
})(SGFTag || (SGFTag = {}));
var NEXT_PLAYER_LABEL = "◍";
var TAG_LABELS = (_a = {},
    _a[SGFTag.Triangle] = "△",
    _a[SGFTag.Square] = "□",
    _a[SGFTag.Circle] = "○",
    _a[SGFTag.X] = "×",
    _a[SGFTag.BlackMove] = NEXT_PLAYER_LABEL,
    _a[SGFTag.WhiteMove] = NEXT_PLAYER_LABEL,
    _a[SGFTag.Label] = "special case",
    _a);
var IS_NIGHT_MODE = !!((_b = document.getElementsByClassName("nightMode")) === null || _b === void 0 ? void 0 : _b.length);
var IS_ANDROID = window.navigator.userAgent.toLowerCase().indexOf("android") > 0;
var bgColor = "#ebb063";
var blackStoneColor = "black";
var whiteStoneColor = "white";
if (IS_NIGHT_MODE && IS_ANDROID) {
    console.log("android in dark mode!");
    bgColor = "#184c96"; // invert of original bgColor
    blackStoneColor = "white";
    whiteStoneColor = "black";
}
function coordsToSgfCoords(row, col) {
    return String.fromCharCode(97 + col) + String.fromCharCode(97 + row);
}
var GobanPosition = /** @class */ (function () {
    function GobanPosition() {
        this.lines = [];
        this.tags = {};
        this.labelsByLocations = {};
        this.whiteTurn = false;
        this.blackTurn = false;
    }
    GobanPosition.prototype.size = function () {
        var _a;
        return ((_a = this.lines) === null || _a === void 0 ? void 0 : _a.length) || 19;
    };
    GobanPosition.prototype.parseLine = function (line) {
        if (line.match(/^[\.wbWB]{5,}$/)) {
            console.log("board line:", line);
            this.lines.push(line);
            if (line.indexOf("B") > 0) {
                this.whiteTurn = true;
            }
            if (line.indexOf("W") >= 0) {
                this.blackTurn = true;
            }
        }
        else if (line.match(/^\d+:.*/)) {
            console.log("diff line:", line);
            var parts = line.split(":");
            this.lines[parseInt(parts[0])] = parts[1];
        }
        else if (line.match(/\w+:.*/)) {
            console.log("tag line:", line);
            var pos = line.indexOf(":");
            var tag = line.substring(0, pos);
            var val = line.substring(pos + 1);
            if (!this.tags[tag]) {
                this.tags[tag] = [];
            }
            console.log("tag: " + tag + ": " + val);
            this.tags[tag].push(val);
            if (tag in TAG_LABELS) {
                for (var _i = 0, _a = val.split(","); _i < _a.length; _i++) {
                    var coords = _a[_i];
                    console.log("tag:" + tag);
                    if (tag === SGFTag.Label) {
                        var _b = coords.split(":"), coord = _b[0], label = _b[1];
                        console.log("label tag:" + tag + "/" + val + "/" + coord + "/" + label);
                        this.labelsByLocations[coord] = label;
                    }
                    else {
                        this.labelsByLocations[coords] = TAG_LABELS[tag];
                    }
                }
                console.log("tags by location");
            }
        }
        else {
            console.error("Invalid line: " + line);
        }
        if (this.tags[SGFTag.WhiteMove]) {
            this.blackTurn = true;
        }
        if (this.tags[SGFTag.BlackMove]) {
            this.whiteTurn = true;
        }
    };
    return GobanPosition;
}());
var Goban = /** @class */ (function () {
    function Goban(sidePx) {
        var _a;
        this.sidePx = sidePx;
        this.positions = [];
        this.drawGoban();
        if ((_a = this.positions) === null || _a === void 0 ? void 0 : _a.length) {
            this.drawBoard(0);
        }
    }
    Goban.prototype.parseGolangPositions = function (content) {
        // fix old version (not using it because veery one of those chars occupy 3 bytes):
        console.log("BEFORE:");
        console.log(content);
        content = content.
            replace(/<div.*?>/g, "\n").
            replace(/<br.*?>/g, "\n").
            replace(/<p.*?>/g, "\n").
            replace(/<.*?>/g, "").
            replace(/·/g, ".").
            replace(/●/g, "b").
            replace(/○/g, "w").
            replace(/↩/g, "\n");
        console.log("AFTER:");
        console.log(content);
        var res = [];
        for (var _i = 0, _a = content.trim().split("\n"); _i < _a.length; _i++) {
            var line = _a[_i];
            //console.log("line:", line);
            if (res.length == 0) {
                res.push(new GobanPosition());
            }
            line = line.trim();
            if (line.indexOf("--") == 0) {
                res.push(new GobanPosition());
                continue;
            }
            if (res.length == 1) { // Initial board:
                res[res.length - 1].parseLine(line);
            }
            else {
                if (res[res.length - 1].lines.length == 0) {
                    for (var _b = 0, _c = res[res.length - 2].lines; _b < _c.length; _b++) {
                        var line_1 = _c[_b];
                        res[res.length - 1].lines.push(line_1.toLowerCase());
                    }
                }
                res[res.length - 1].parseLine(line);
            }
        }
        return res;
    };
    Goban.prototype.drawGoban = function () {
        this.gobanEl = document.getElementById("goban");
        this.positions = this.parseGolangPositions(this.gobanEl.innerHTML.trim());
        this.boardSize = this.positions[0].size();
        console.log("board size: " + this.boardSize);
        this.bandWitdh = this.sidePx / (this.boardSize - 1);
        this.stoneSide = this.bandWitdh * 0.95;
        this.containerDiv = document.createElement("div");
        this.containerDiv.style.position = "relative";
        this.containerDiv.style.backgroundColor = bgColor;
        this.containerDiv.style.border = "0.01px solid gray";
        this.containerDiv.style.width = this.sidePx + this.bandWitdh * 2 + "px";
        this.containerDiv.style.height = this.sidePx + this.bandWitdh * 2 + "px";
        this.gobanLinesDiv = document.createElement("div");
        this.gobanLinesDiv.style.position = "absolute";
        this.gobanLinesDiv.style.width = this.sidePx + "px";
        this.gobanLinesDiv.style.height = this.sidePx + "px";
        this.gobanLinesDiv.style.left = this.bandWitdh + "px";
        this.gobanLinesDiv.style.top = this.bandWitdh + "px";
        this.gobanLinesDiv.style.backgroundColor = bgColor;
        this.containerDiv.appendChild(this.gobanLinesDiv);
        for (var i = 0; i < this.boardSize; i++) {
            for (var j = 0; j < 2; j++) {
                var lineDiv = document.createElement("div");
                lineDiv.style.border = "0.5px solid black";
                lineDiv.style.position = "absolute";
                lineDiv.style.borderWidth = "1px 1px 0px 0px";
                if (j == 0) {
                    lineDiv.style.width = "1px";
                    lineDiv.style.height = this.sidePx + "px";
                    lineDiv.style.left = i * this.bandWitdh + "px";
                    lineDiv.style.top = 0 + "px";
                }
                else {
                    lineDiv.style.width = this.sidePx + "px";
                    lineDiv.style.height = "1px";
                    lineDiv.style.top = i * this.bandWitdh + "px";
                    lineDiv.style.left = 0 + "px";
                }
                this.gobanLinesDiv.appendChild(lineDiv);
            }
        }
        this.gobanEl.innerHTML = "";
        this.gobanEl.appendChild(this.containerDiv);
    };
    Goban.prototype.drawBoard = function (position) {
        this.drawStones(this.positions[position % this.positions.length]);
    };
    Goban.prototype.drawStones = function (g) {
        var _a, _b;
        for (var col = 0; col < this.boardSize; col++) {
            for (var row = 0; row < this.boardSize; row++) {
                this.drawStone(g, row, col);
            }
        }
        var turnEl = document.getElementById("goban_turn");
        if (turnEl) {
            if (g.blackTurn) {
                turnEl.innerHTML = "<strong>BLACK<strong> to play";
            }
            else if (g.whiteTurn) {
                turnEl.innerHTML = "<strong>WHITE</strong> to play";
            }
            else {
                turnEl.innerHTML = "";
            }
        }
        var commentsEl = document.getElementById("goban_comment");
        if (commentsEl) {
            commentsEl.innerHTML = ((_b = (_a = g.tags[SGFTag.Comment]) === null || _a === void 0 ? void 0 : _a.map(function (el) { return el.split("\\n").join("<br/>"); })) === null || _b === void 0 ? void 0 : _b.join("<br/>")) || "";
        }
    };
    Goban.prototype.drawStone = function (g, row, column) {
        var id = "stone-" + row + "-" + column;
        var existingDiv = document.getElementById(id);
        var stoneDiv = existingDiv || document.createElement("div");
        if (!existingDiv) {
            stoneDiv.id = id;
            stoneDiv.style.position = "absolute";
            stoneDiv.style.textAlign = "center";
            stoneDiv.style.left = (1 + column * this.bandWitdh - 0.5 * this.stoneSide) + "px";
            stoneDiv.style.top = (1 + row * this.bandWitdh - 0.5 * this.stoneSide) + "px";
            stoneDiv.style.width = this.stoneSide + "px";
            stoneDiv.style.height = this.stoneSide + "px";
            stoneDiv.onclick = function () {
                alert("Location " + coordsToSgfCoords(row, column));
            };
            this.gobanLinesDiv.appendChild(stoneDiv);
        }
        stoneDiv.innerHTML = "";
        var stone = (g.lines[row] || [])[column] || MIDDOT;
        switch (stone.toLowerCase()) {
            case MIDDOT:
                stoneDiv.style.backgroundColor = null;
                break;
            case BLACK_CIRCLE:
                stoneDiv.style.backgroundColor = blackStoneColor;
                break;
            case WHITE_CIRCLE:
                stoneDiv.style.backgroundColor = whiteStoneColor;
                break;
        }
        var label = g.labelsByLocations[coordsToSgfCoords(row, column)];
        var isLatestMove = stone == "W" || stone == "B";
        if (label || isLatestMove) {
            var centerDiv = document.createElement("div");
            if (isLatestMove) {
                centerDiv.innerHTML = NEXT_PLAYER_LABEL;
            }
            else {
                centerDiv.innerHTML = label;
            }
            centerDiv.style.position = "absolute";
            centerDiv.style.color = "gray";
            centerDiv.style.left = "50%";
            centerDiv.style.top = "50%";
            centerDiv.style.transform = "translate(-50%, -55%)";
            centerDiv.style.fontSize = this.stoneSide * 0.75 + "px";
            centerDiv.style.textAlign = "center";
            stoneDiv.appendChild(centerDiv);
        }
        stoneDiv.style.borderRadius = this.stoneSide * 0.5 + "px";
    };
    Goban.prototype.animate = function (initDelay, interval) {
        var _this = this;
        this.drawBoard(0);
        var _loop_1 = function (n) {
            (function (pos) {
                setTimeout(function () {
                    _this.drawBoard(n);
                }, n == 0 ? initDelay : initDelay + (pos - 1) * interval);
            })(n);
        };
        for (var n = 0; n < this.positions.length; n++) {
            _loop_1(n);
        }
    };
    return Goban;
}());

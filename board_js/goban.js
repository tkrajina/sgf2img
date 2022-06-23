var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
var __generator = (this && this.__generator) || function (thisArg, body) {
    var _ = { label: 0, sent: function() { if (t[0] & 1) throw t[1]; return t[1]; }, trys: [], ops: [] }, f, y, t, g;
    return g = { next: verb(0), "throw": verb(1), "return": verb(2) }, typeof Symbol === "function" && (g[Symbol.iterator] = function() { return this; }), g;
    function verb(n) { return function (v) { return step([n, v]); }; }
    function step(op) {
        if (f) throw new TypeError("Generator is already executing.");
        while (_) try {
            if (f = 1, y && (t = op[0] & 2 ? y["return"] : op[0] ? y["throw"] || ((t = y["return"]) && t.call(y), 0) : y.next) && !(t = t.call(y, op[1])).done) return t;
            if (y = 0, t) op = [op[0] & 2, t.value];
            switch (op[0]) {
                case 0: case 1: t = op; break;
                case 4: _.label++; return { value: op[1], done: false };
                case 5: _.label++; y = op[1]; op = [0]; continue;
                case 7: op = _.ops.pop(); _.trys.pop(); continue;
                default:
                    if (!(t = _.trys, t = t.length > 0 && t[t.length - 1]) && (op[0] === 6 || op[0] === 2)) { _ = 0; continue; }
                    if (op[0] === 3 && (!t || (op[1] > t[0] && op[1] < t[3]))) { _.label = op[1]; break; }
                    if (op[0] === 6 && _.label < t[1]) { _.label = t[1]; t = op; break; }
                    if (t && _.label < t[2]) { _.label = t[2]; _.ops.push(op); break; }
                    if (t[2]) _.ops.pop();
                    _.trys.pop(); continue;
            }
            op = body.call(thisArg, _);
        } catch (e) { op = [6, e]; y = 0; } finally { f = t = 0; }
        if (op[0] & 5) throw op[1]; return { value: op[0] ? op[1] : void 0, done: true };
    }
};
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
var NEXT_PLAYER_LABEL = "●";
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
        this.whitePlays = false;
        this.blackPlays = false;
    }
    GobanPosition.prototype.size = function () {
        var _a;
        return ((_a = this.lines) === null || _a === void 0 ? void 0 : _a.length) || 19;
    };
    GobanPosition.prototype.parseLine = function (line) {
        if (line.match(/^[\.wbWB]{5,}$/)) {
            console.log("board line:", line);
            this.lines.push(line);
            this.whitePlays = line.indexOf("B") >= 0;
            this.blackPlays = line.indexOf("W") >= 0;
        }
        else if (line.match(/^\d+:.*/)) {
            console.log("diff line:", line);
            var parts = line.split(":");
            this.lines[parseInt(parts[0])] = parts[1];
            this.whitePlays = line.indexOf("B") >= 0;
            this.blackPlays = line.indexOf("W") >= 0;
        }
        else if (line.toLocaleLowerCase().match(/crop:.*/)) {
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
            this.blackPlays = true;
        }
        if (this.tags[SGFTag.BlackMove]) {
            this.whitePlays = true;
        }
    };
    return GobanPosition;
}());
var Goban = /** @class */ (function () {
    function Goban(sidePx) {
        var _a;
        this.sidePx = sidePx;
        this.positions = [];
        this.cropTop = 0;
        this.cropRight = 0;
        this.cropBottom = 0;
        this.cropLeft = 0;
        this.position = 0;
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
            replace(/<.*?>/g, "");
        console.log("AFTER:");
        console.log(content);
        var res = [];
        for (var _i = 0, _a = content.trim().split("\n"); _i < _a.length; _i++) {
            var line = _a[_i];
            //console.log("line:", line);
            if (line.trim().toLowerCase().match(/^crop:.*/)) {
                var parts = line.split(":")[1].trim().split(/[\s,]+/) || ["0", "0", "0", "0"];
                this.cropTop = parseFloat(parts[0]) || 0;
                this.cropRight = parseFloat(parts[1]) || 0;
                this.cropBottom = parseFloat(parts[2]) || 0;
                this.cropLeft = parseFloat(parts[3]) || 0;
                continue;
            }
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
        this.containerElement = document.getElementById("goban");
        this.positions = this.parseGolangPositions(this.containerElement.innerHTML.trim());
        this.boardSize = this.positions[0].size();
        console.log("board size: " + this.boardSize);
        this.bandWitdh = this.sidePx / (this.boardSize - 1);
        this.stoneSide = this.bandWitdh * 0.95;
        var containerWindowDiv = document.createElement("div");
        containerWindowDiv.style.position = "relative";
        containerWindowDiv.style.overflow = "hidden";
        //containerWindowDiv.style.border = "5px solid red";
        containerWindowDiv.style.width = (1 - this.cropRight - this.cropLeft) * (this.sidePx + this.bandWitdh * 2) + "px";
        containerWindowDiv.style.height = (1 - this.cropBottom - this.cropTop) * (this.sidePx + this.bandWitdh * 2) + "px";
        this.gobanDiv = document.createElement("div");
        this.gobanDiv.style.position = "absolute";
        this.gobanDiv.style.top = (-this.cropTop) * (this.sidePx + this.bandWitdh * 2) + "px";
        this.gobanDiv.style.left = (-this.cropLeft) * (this.sidePx + this.bandWitdh * 2) + "px";
        this.gobanDiv.style.overflow = "hidden";
        this.gobanDiv.style.marginBottom = -50 + "px";
        this.gobanDiv.style.backgroundColor = bgColor;
        this.gobanDiv.style.border = "0.01px solid gray";
        this.gobanDiv.style.width = this.sidePx + this.bandWitdh * 2 + "px";
        this.gobanDiv.style.height = this.sidePx + this.bandWitdh * 2 + "px";
        containerWindowDiv.appendChild(this.gobanDiv);
        var gobanLinesDiv = document.createElement("div");
        gobanLinesDiv.style.position = "absolute";
        gobanLinesDiv.style.width = this.sidePx + "px";
        gobanLinesDiv.style.height = this.sidePx + "px";
        gobanLinesDiv.style.left = this.bandWitdh + "px";
        gobanLinesDiv.style.top = this.bandWitdh + "px";
        gobanLinesDiv.style.backgroundColor = bgColor;
        this.gobanDiv.appendChild(gobanLinesDiv);
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
                gobanLinesDiv.appendChild(lineDiv);
            }
        }
        this.containerElement.innerHTML = "";
        this.containerElement.appendChild(containerWindowDiv);
        this.drawHoshi();
    };
    Goban.prototype.drawBoard = function (position) {
        if ("number" === typeof position) {
            this.position = position;
        }
        if (this.position >= this.positions.length - 1) {
            this.stopAnimation();
        }
        this.position = this.position % this.positions.length;
        if (this.position < 0) {
            this.position += this.positions.length;
        }
        var el = document.getElementById("goban_position");
        if (el) {
            el.innerHTML = this.position + 1 + "/" + this.positions.length;
        }
        this.drawStones(this.positions[this.position], this.positions[this.position + 1]);
    };
    Goban.prototype.drawHoshi = function () {
        var hoshiRadious = this.stoneSide / 4;
        var hoshiPositions = [
            [3, 3],
            [3, 9],
            [3, 15],
            [9, 3],
            [9, 9],
            [9, 15],
            [15, 3],
            [15, 9],
            [15, 15],
        ];
        for (var _i = 0, hoshiPositions_1 = hoshiPositions; _i < hoshiPositions_1.length; _i++) {
            var pos = hoshiPositions_1[_i];
            var row = pos[0], column = pos[1];
            var id = "hoshi-" + row + "-" + column;
            var hoshiDiv = document.createElement("div");
            hoshiDiv.id = id;
            hoshiDiv.style.position = "absolute";
            hoshiDiv.style.textAlign = "center";
            hoshiDiv.style.left = 1.5 + (1 + column) * this.bandWitdh - 0.5 * hoshiRadious + "px";
            hoshiDiv.style.top = 0.5 + (1 + row) * this.bandWitdh - 0.5 * hoshiRadious + "px";
            hoshiDiv.style.width = hoshiRadious + "px";
            hoshiDiv.style.height = hoshiRadious + "px";
            hoshiDiv.style.backgroundColor = blackStoneColor;
            hoshiDiv.style.borderRadius = hoshiRadious * 0.5 + "px";
            this.gobanDiv.appendChild(hoshiDiv);
        }
    };
    Goban.prototype.drawStones = function (g, next) {
        var _a, _b;
        for (var col = 0; col < this.boardSize; col++) {
            for (var row = 0; row < this.boardSize; row++) {
                this.drawStone(g, row, col);
            }
        }
        var turnEl = document.getElementById("goban_turn");
        if (turnEl) {
            var radius = 20;
            var padding = .25;
            turnEl.innerHTML = "";
            var nextStoneBgDiv = document.createElement("div");
            var nextStoneDiv = document.createElement("div");
            nextStoneBgDiv.style.backgroundColor = bgColor;
            nextStoneBgDiv.style.width = radius * (1 + padding) + "px";
            nextStoneBgDiv.style.height = radius * (1 + padding) + "px";
            nextStoneBgDiv.style.color = blackStoneColor;
            nextStoneBgDiv.style.position = "relative";
            nextStoneDiv.style.borderRadius = (radius / 2) + "px";
            nextStoneDiv.style.position = "absolute";
            nextStoneDiv.style.top = (radius * padding / 2) + "px";
            nextStoneDiv.style.left = (radius * padding / 2) + "px";
            nextStoneDiv.style.width = radius + "px";
            nextStoneDiv.style.height = radius + "px";
            if (next === null || next === void 0 ? void 0 : next.blackPlays) {
                nextStoneDiv.style.backgroundColor = blackStoneColor;
                nextStoneBgDiv.appendChild(nextStoneDiv);
            }
            else if (next === null || next === void 0 ? void 0 : next.whitePlays) {
                nextStoneDiv.style.backgroundColor = whiteStoneColor;
                nextStoneBgDiv.appendChild(nextStoneDiv);
            }
            turnEl.appendChild(nextStoneBgDiv);
        }
        var commentsEl = document.getElementById("goban_comment");
        console.log("draw with comment" + g.tags[SGFTag.Comment]);
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
            stoneDiv.style.left = (1 + column) * this.bandWitdh - 0.5 * this.stoneSide + "px";
            stoneDiv.style.top = (1 + row) * this.bandWitdh - 0.5 * this.stoneSide + "px";
            stoneDiv.style.width = this.stoneSide + "px";
            stoneDiv.style.height = this.stoneSide + "px";
            stoneDiv.onclick = function () {
                alert("Location " + coordsToSgfCoords(row, column));
            };
            this.gobanDiv.appendChild(stoneDiv);
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
            switch (stone.toLowerCase()) {
                case "w":
                    centerDiv.style.color = blackStoneColor;
                    break;
                case "b":
                    centerDiv.style.color = whiteStoneColor;
                    break;
                default:
                    centerDiv.style.color = blackStoneColor;
            }
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
        this.stopAnimation();
        this.drawBoard(0);
        var n = 0;
        this.animationTimeout = setTimeout(function () {
            _this.drawBoard(++n);
            if (n >= _this.positions.length - 1) {
                return;
            }
            _this.animationInterval = setInterval(function () {
                _this.drawBoard(++n);
            }, interval);
        }, initDelay);
    };
    Goban.prototype.stopAnimation = function () {
        clearTimeout(this.animationTimeout);
        clearInterval(this.animationInterval);
    };
    Goban.prototype.next = function () {
        this.stopAnimation();
        this.drawBoard(this.position + 1);
    };
    Goban.prototype.previous = function () {
        this.stopAnimation();
        this.drawBoard(this.position - 1);
    };
    Goban.prototype.first = function () {
        this.stopAnimation();
        this.drawBoard(0);
    };
    Goban.prototype.last = function () {
        this.stopAnimation();
        this.drawBoard(this.positions.length - 1);
    };
    Goban.prototype.sgf = function () {
        return __awaiter(this, void 0, void 0, function () {
            var sgf, n, pos, tag, _i, _a, val, lineNo, line, columnNo, loc, e_1, commentsEl, textarea;
            return __generator(this, function (_b) {
                switch (_b.label) {
                    case 0:
                        sgf = "(";
                        for (n = 0; n < this.positions.length; n++) {
                            if (n > 0) {
                                sgf += "\n;";
                            }
                            pos = this.positions[n];
                            for (tag in pos.tags) {
                                for (_i = 0, _a = pos.tags[tag]; _i < _a.length; _i++) {
                                    val = _a[_i];
                                    sgf += "\n" + (tag + "[" + val + "]");
                                }
                            }
                            for (lineNo = 0; lineNo < pos.lines.length; lineNo++) {
                                line = pos.lines[lineNo].trim();
                                console.log("line:", line);
                                for (columnNo = 0; columnNo < line.length; columnNo++) {
                                    loc = line[columnNo];
                                    switch (loc) {
                                        case "w":
                                            if (n == 0) {
                                                sgf += "\n" + ("AW[" + this.toSgfCoordinates(lineNo, columnNo) + "]");
                                            }
                                            break;
                                        case "W":
                                            if (!pos.tags["W"]) {
                                                sgf += "\n" + ("W[" + this.toSgfCoordinates(lineNo, columnNo) + "]");
                                            }
                                            break;
                                        case "b":
                                            if (n == 0) {
                                                sgf += "\n" + ("AB[" + this.toSgfCoordinates(lineNo, columnNo) + "]");
                                            }
                                            break;
                                        case "B":
                                            if (!pos.tags["B"]) {
                                                sgf += "\n" + ("B[" + this.toSgfCoordinates(lineNo, columnNo) + "]");
                                            }
                                            break;
                                    }
                                }
                            }
                        }
                        sgf += "\n)";
                        _b.label = 1;
                    case 1:
                        _b.trys.push([1, 3, , 4]);
                        return [4 /*yield*/, navigator.clipboard.writeText(sgf)];
                    case 2:
                        _b.sent();
                        alert("Copied to clipboard");
                        return [3 /*break*/, 4];
                    case 3:
                        e_1 = _b.sent();
                        commentsEl = document.getElementById("goban_comment");
                        commentsEl.innerHTML = "SGF:<br/>";
                        textarea = document.createElement("textarea");
                        textarea.value = sgf;
                        commentsEl.appendChild(textarea);
                        textarea.select();
                        return [3 /*break*/, 4];
                    case 4: return [2 /*return*/];
                }
            });
        });
    };
    Goban.prototype.toSgfCoordinates = function (lineNo, columnNo) {
        var coords = "abcdefghijklmnopqrs";
        return coords[columnNo] + coords[lineNo];
    };
    return Goban;
}());

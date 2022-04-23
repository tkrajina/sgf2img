const MIDDOT = ".";
const BLACK_CIRCLE = "b";
const WHITE_CIRCLE = "w";

enum SGFTag {
	AddBlack    = "AB",
	AddWhite    = "AW",
	Annotations = "AN",
	Application = "AP",
	BlackMove   = "B",
	BlackRank   = "BR",
	BlackTeam   = "BT",
	Comment     = "C",
	Copyright   = "CP",
	Date        = "DT",
	Event       = "EV",
	FileFormat  = "FF",
	Game        = "GM",
	GameName    = "GN",
	Handicap    = "HA",
	Komi        = "KM",
	Opening     = "ON",
	Overtime    = "OT",
	BlackName   = "PB",
	Place       = "PC",
	Player      = "PL",
	WhiteName   = "PW",
	Result      = "RE",
	Round       = "RO",
	Rules       = "RU",
	Source      = "SO",
	Size        = "SZ",
	Timelimit   = "TM",
	User        = "US",
	WhiteMove   = "W",
	WhiteRank   = "WR",
	WhiteTeam   = "WT",

	// Additional
	Triangle = "TR",
	Square   = "SQ",
	Circle   = "CR",
	X        = "MA",
	Label    = "LB",
}

type Stone = '.' | 'B' | 'W' | 'b' | 'w';

const NEXT_PLAYER_LABEL = "●";

const TAG_LABELS = {
	[SGFTag.Triangle]: "△",
	[SGFTag.Square]: "□",
	[SGFTag.Circle]: "○",
	[SGFTag.X]: "×",
	[SGFTag.BlackMove]: NEXT_PLAYER_LABEL,
	[SGFTag.WhiteMove]: NEXT_PLAYER_LABEL,
	[SGFTag.Label]: "special case", // overwritten later
};

const IS_NIGHT_MODE = !!document.getElementsByClassName("nightMode")?.length;
const IS_ANDROID = window.navigator.userAgent.toLowerCase().indexOf("android") > 0

let bgColor = "#ebb063";
let blackStoneColor = "black"
let whiteStoneColor = "white"

if (IS_NIGHT_MODE && IS_ANDROID) {
	console.log("android in dark mode!")
	bgColor = "#184c96"; // invert of original bgColor
	blackStoneColor = "white";
	whiteStoneColor = "black";
}

function coordsToSgfCoords(row: number, col: number) {
	return String.fromCharCode(97+col) + String.fromCharCode(97+row);
}

class GobanPosition {
	lines: string[] = [];
	tags: {[tag: string]: string[]} = {};
	labelsByLocations: {[coordinates: string]: string} = {};
	whiteTurn = false;
	blackTurn = false;

	constructor() {
	}

	size() {
		return this.lines?.length || 19;
	}

	parseLine(line: string) {
		if (line.match(/^[\.wbWB]{5,}$/)) {
			console.log("board line:", line);
			this.lines.push(line);
			if (line.indexOf("B") > 0) {
				this.whiteTurn = true
			}
			if (line.indexOf("W") >= 0) {
				this.blackTurn = true
			}
		} else if (line.match(/^\d+:.*/)) {
			console.log("diff line:", line);
			const parts = line.split(":");
			this.lines[parseInt(parts[0])] = parts[1]
		} else if (line.match(/\w+:.*/)) {
			console.log("tag line:", line);
			const pos = line.indexOf(":");
			const tag = line.substring(0, pos);
			const val = line.substring(pos + 1);
			if (!this.tags[tag]) {
				this.tags[tag] = [];
			}
			console.log(`tag: ${tag}: ${val}`);
			this.tags[tag].push(val);

			if (tag in TAG_LABELS) {
				for (const coords of val.split(",")) {
					console.log("tag:" + tag);
					if (tag === SGFTag.Label) {
						let [coord, label] = coords.split(":");
						console.log("label tag:" + tag + "/" + val + "/" + coord + "/" + label);
						this.labelsByLocations[coord] = label;
					} else {
						this.labelsByLocations[coords] = TAG_LABELS[tag];
					}
				}
				console.log("tags by location")
			}
		} else {
			console.error("Invalid line: " + line);
		}

		if (this.tags[SGFTag.WhiteMove]) {
			this.blackTurn = true;
		}
		if (this.tags[SGFTag.BlackMove]) {
			this.whiteTurn = true;
		}

	}
}

class Goban {

	private gobanEl: HTMLElement
	private containerDiv: HTMLElement;
	private gobanLinesDiv: HTMLElement;

	positions: GobanPosition[] = [];

	boardSize: number;
	bandWitdh: number;
	stoneSide: number;

	constructor(private readonly sidePx: number) {
		this.drawGoban();
		if (this.positions?.length) {
			this.drawBoard(0);
		}
	}

	private parseGolangPositions(content: string) {
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
		const res: GobanPosition[] = [];
		for (let line of content.trim().split("\n")) {
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
			} else {
				if (res[res.length - 1].lines.length == 0) {
					for (const line of res[res.length - 2].lines) {
						res[res.length - 1].lines.push(line.toLowerCase());
					}
				}
				res[res.length - 1].parseLine(line);
			}
		}
		return res;
	}

	drawGoban() {
		this.gobanEl = document.getElementById("goban")
		this.positions = this.parseGolangPositions(this.gobanEl.innerHTML.trim());

		this.boardSize = this.positions[0].size()
		console.log(`board size: ${this.boardSize}`);
		this.bandWitdh = this.sidePx / (this.boardSize - 1);
		this.stoneSide = this.bandWitdh * 0.95;


		this.containerDiv = document.createElement("div");
		this.containerDiv.style.position = "relative";
		this.containerDiv.style.backgroundColor = bgColor;
		this.containerDiv.style.border = "0.01px solid gray";
		this.containerDiv.style.width = `${this.sidePx + this.bandWitdh*2}px`;
		this.containerDiv.style.height = `${this.sidePx + this.bandWitdh*2}px`;

		this.gobanLinesDiv = document.createElement("div");
		this.gobanLinesDiv.style.position = "absolute";
		this.gobanLinesDiv.style.width = `${this.sidePx}px`;
		this.gobanLinesDiv.style.height = `${this.sidePx}px`;
		this.gobanLinesDiv.style.left = `${this.bandWitdh}px`;
		this.gobanLinesDiv.style.top = `${this.bandWitdh}px`;
		this.gobanLinesDiv.style.backgroundColor = bgColor;
		this.containerDiv.appendChild(this.gobanLinesDiv);

		for (let i = 0; i < this.boardSize; i++) {
			for (let j = 0; j < 2; j++) {
				const lineDiv = document.createElement("div");
				lineDiv.style.border = "0.5px solid black";
				lineDiv.style.position = "absolute";
				lineDiv.style.borderWidth = "1px 1px 0px 0px";
				if (j == 0) {
					lineDiv.style.width = `1px`;
					lineDiv.style.height = `${this.sidePx}px`;
					lineDiv.style.left = `${i*this.bandWitdh}px`;
					lineDiv.style.top = `${0}px`;
				} else {
					lineDiv.style.width = `${this.sidePx}px`;
					lineDiv.style.height = `1px`;
					lineDiv.style.top = `${i*this.bandWitdh}px`;
					lineDiv.style.left = `${0}px`;
				}
				this.gobanLinesDiv.appendChild(lineDiv);
			}
		}

		this.gobanEl.innerHTML = "";
		this.gobanEl.appendChild(this.containerDiv);
	}

	drawBoard(position: number) {
		this.drawStones(this.positions[position % this.positions.length]);
	}

	drawStones(g: GobanPosition) {
		for (let col = 0; col < this.boardSize; col++) {
			for (let row = 0; row < this.boardSize; row++) {
				this.drawStone(g, row, col);
			}
		}
		let turnEl = document.getElementById("goban_turn");
		if (turnEl) {
			if (g.blackTurn) {
				turnEl.innerHTML = "<strong>BLACK<strong> to play";
			} else if (g.whiteTurn) {
				turnEl.innerHTML = "<strong>WHITE</strong> to play";
			} else {
				turnEl.innerHTML = "";
			}
		}
		let commentsEl = document.getElementById("goban_comment");
		if (commentsEl) {
			commentsEl.innerHTML = g.tags[SGFTag.Comment]?.map(el => el.split("\\n").join("<br/>"))?.join("<br/>") || "";
		}
	}

	drawStone(g: GobanPosition, row: number, column: number) {
		const id = `stone-${row}-${column}`;
		const existingDiv = document.getElementById(id)
		const stoneDiv = existingDiv || document.createElement("div");
		if (!existingDiv) {
			stoneDiv.id = id;
			stoneDiv.style.position = "absolute";
			stoneDiv.style.textAlign = "center";
			stoneDiv.style.left = `${(1 + column * this.bandWitdh - 0.5 * this.stoneSide)}px`;
			stoneDiv.style.top = `${(1 + row * this.bandWitdh - 0.5 * this.stoneSide)}px`;
			stoneDiv.style.width = `${this.stoneSide}px`;
			stoneDiv.style.height = `${this.stoneSide}px`;
			stoneDiv.onclick = () => {
				alert("Location " + coordsToSgfCoords(row, column));

			}
			this.gobanLinesDiv.appendChild(stoneDiv);
		}
		stoneDiv.innerHTML = "";

		const stone = (g.lines[row] || [])[column] || MIDDOT;
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

		const label = g.labelsByLocations[coordsToSgfCoords(row, column)];
		const isLatestMove = stone == "W" || stone == "B";
		if (label || isLatestMove) {
			const centerDiv = document.createElement("div");
			if (isLatestMove) {
				centerDiv.innerHTML = NEXT_PLAYER_LABEL;
			} else {
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
				centerDiv.style.color = "red";
			}
			centerDiv.style.left = "50%";
			centerDiv.style.top = "50%";
			centerDiv.style.transform = "translate(-50%, -55%)";
			centerDiv.style.fontSize = `${this.stoneSide * 0.75}px`;
			centerDiv.style.textAlign = "center";
			stoneDiv.appendChild(centerDiv);
		}

		stoneDiv.style.borderRadius = `${this.stoneSide * 0.5}px`;
	}

	public animate(initDelay?: number, interval?: number) {
		this.drawBoard(0);
		for (let n = 0; n < this.positions.length; n++) {
			((pos: number) => {
				setTimeout(() => {
					this.drawBoard(n);
				}, n == 0 ? initDelay : initDelay + (pos - 1) * interval);
			})(n);
		}
	}
}
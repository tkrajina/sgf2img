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
	whiteNext = false;
	blackNext = false;

	constructor() {
	}

	size() {
		return this.lines?.length || 19;
	}

	parseLine(line: string) {
		if (line.match(/^[\.wbWB]{5,}$/)) {
			console.log("board line:", line);
			this.lines.push(line);
			if (line.indexOf("B") >= 0) this.whiteNext = true;
			if (line.indexOf("W") >= 0) this.blackNext = true;
		} else if (line.match(/^\d+:.*/)) {
			console.log("diff line:", line);
			const parts = line.split(":");
			this.lines[parseInt(parts[0])] = parts[1]
			if (line.indexOf("B") >= 0) this.whiteNext = true;
			if (line.indexOf("W") >= 0) this.blackNext = true;
		} else if (line.toLocaleLowerCase().match(/crop:.*/)) {
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
			this.blackNext = true;
		}
		if (this.tags[SGFTag.BlackMove]) {
			this.whiteNext = true;
		}

	}
}

class Goban {

	private containerElement: HTMLElement
	private gobanDiv: HTMLElement;

	positions: GobanPosition[] = [];

	boardSize: number;
	bandWitdh: number;
	stoneSide: number;

	cropTop = 0;
	cropRight = 0;
	cropBottom = 0;
	cropLeft = 0;

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
			replace(/<.*?>/g, "");
		console.log("AFTER:");
		console.log(content);
		const res: GobanPosition[] = [new GobanPosition()];
		for (let line of content.trim().split("\n")) {
			//console.log("line:", line);
			if (line.trim().toLowerCase().match(/^crop:.*/)) {
				const parts = line.split(":")[1].trim().split(/[\s,]+/) || ["0","0","0","0"];
				this.cropTop = parseFloat(parts[0]) || 0;
				this.cropRight = parseFloat(parts[1]) || 0;
				this.cropBottom = parseFloat(parts[2]) || 0;
				this.cropLeft = parseFloat(parts[3]) || 0;
				continue;
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

		// If first position has no last move:
		if (res.length > 1 && !res[0].whiteNext && !res[0].blackNext) {
			res[0].blackNext = !res[1].blackNext
			res[0].whiteNext = !res[1].whiteNext
		}
	
		return res;
	}

	drawGoban() {
		this.containerElement = document.getElementById("goban")
		this.positions = this.parseGolangPositions(this.containerElement.innerHTML.trim());

		this.boardSize = this.positions[0].size()
		console.log(`board size: ${this.boardSize}`);
		this.bandWitdh = this.sidePx / (this.boardSize - 1);
		this.stoneSide = this.bandWitdh * 0.95;

		const containerWindowDiv = document.createElement("div")
		containerWindowDiv.style.position = "relative";
		containerWindowDiv.style.overflow = "hidden";
		//containerWindowDiv.style.border = "5px solid red";
		containerWindowDiv.style.width = `${(1 - this.cropRight - this.cropLeft) * (this.sidePx + this.bandWitdh*2)}px`;
		containerWindowDiv.style.height = `${(1 - this.cropBottom - this.cropTop) * (this.sidePx + this.bandWitdh*2)}px`;

		this.gobanDiv = document.createElement("div");
		this.gobanDiv.style.position = "absolute";
		this.gobanDiv.style.top = `${(- this.cropTop) * (this.sidePx + this.bandWitdh*2)}px`;
		this.gobanDiv.style.left = `${(- this.cropLeft) * (this.sidePx + this.bandWitdh*2)}px`;
		this.gobanDiv.style.overflow = "hidden";
		this.gobanDiv.style.marginBottom = `${-50}px`;
		this.gobanDiv.style.backgroundColor = bgColor;
		this.gobanDiv.style.border = "0.01px solid gray";
		this.gobanDiv.style.width = `${this.sidePx + this.bandWitdh*2}px`;
		this.gobanDiv.style.height = `${this.sidePx + this.bandWitdh*2}px`;
		containerWindowDiv.appendChild(this.gobanDiv);

		const gobanLinesDiv = document.createElement("div");
		gobanLinesDiv.style.position = "absolute";
		gobanLinesDiv.style.width = `${this.sidePx}px`;
		gobanLinesDiv.style.height = `${this.sidePx}px`;
		gobanLinesDiv.style.left = `${this.bandWitdh}px`;
		gobanLinesDiv.style.top = `${this.bandWitdh}px`;
		gobanLinesDiv.style.backgroundColor = bgColor;
		this.gobanDiv.appendChild(gobanLinesDiv);

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
				gobanLinesDiv.appendChild(lineDiv);
			}
		}

		this.containerElement.innerHTML = "";
		this.containerElement.appendChild(containerWindowDiv);
		this.drawHoshi();
	}

	drawBoard(position?: number) {
		if ("number" === typeof position) {
			this.position = position as number;
		}
		if (this.position >= this.positions.length - 1) {
			this.stopAnimation();
		}
		this.position = this.position % this.positions.length
		if (this.position < 0) {
			this.position += this.positions.length;
		}
		const el = document.getElementById("goban_position")
		if (el) {
			el.innerHTML = `${this.position + 1}/${this.positions.length}`;
		}
		this.drawStones(this.positions[this.position]);
	}

	private drawHoshi() {
		const hoshiRadious = this.stoneSide / 4;
		let hoshiPositions: [number, number][]  = [
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
		for (const pos of hoshiPositions) {
			let row = pos[0], column = pos[1];
			const id = `hoshi-${row}-${column}`;
			const hoshiDiv = document.createElement("div");
			hoshiDiv.id = id;
			hoshiDiv.style.position = "absolute";
			hoshiDiv.style.textAlign = "center";
			hoshiDiv.style.left = `${1.5 + (1 + column) * this.bandWitdh - 0.5 * hoshiRadious}px`;
			hoshiDiv.style.top = `${0.5 + (1 + row) * this.bandWitdh - 0.5 * hoshiRadious}px`;
			hoshiDiv.style.width = `${hoshiRadious}px`;
			hoshiDiv.style.height = `${hoshiRadious}px`;
			hoshiDiv.style.backgroundColor = blackStoneColor;
			hoshiDiv.style.borderRadius = `${hoshiRadious * 0.5}px`;
			this.gobanDiv.appendChild(hoshiDiv);
		}
	}

	private drawStones(g: GobanPosition) {
		for (let col = 0; col < this.boardSize; col++) {
			for (let row = 0; row < this.boardSize; row++) {
				this.drawStone(g, row, col);
			}
		}
		let turnEl = document.getElementById("goban_turn");
		if (turnEl) {
			const radius = 20;
			const padding = .25
			turnEl.innerHTML = "";
			const nextStoneBgDiv = document.createElement("div");
			const nextStoneDiv = document.createElement("div");
			nextStoneBgDiv.style.backgroundColor = bgColor;
			nextStoneBgDiv.style.width = `${radius * (1+padding)}px`;
			nextStoneBgDiv.style.height = `${radius * (1 + padding)}px`;
			nextStoneBgDiv.style.color = blackStoneColor;
			nextStoneBgDiv.style.position = "relative";

			nextStoneDiv.style.borderRadius = (radius / 2) + "px";
			nextStoneDiv.style.position = "absolute";
			nextStoneDiv.style.top = `${(radius * padding / 2)}px`;
			nextStoneDiv.style.left = `${(radius * padding / 2)}px`;
			nextStoneDiv.style.width = `${radius}px`;
			nextStoneDiv.style.height = `${radius}px`;

			if (g.blackNext) {
				nextStoneDiv.style.backgroundColor = blackStoneColor;
				nextStoneBgDiv.appendChild(nextStoneDiv);
			} else if (g.whiteNext) {
				nextStoneDiv.style.backgroundColor = whiteStoneColor;
				nextStoneBgDiv.appendChild(nextStoneDiv);
			}
			turnEl.appendChild(nextStoneBgDiv);
		}
		let commentsEl = document.getElementById("goban_comment");
		console.log("draw with comment" + g.tags[SGFTag.Comment]);
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
			stoneDiv.style.left = `${(1 + column) * this.bandWitdh - 0.5 * this.stoneSide}px`;
			stoneDiv.style.top = `${(1 + row) * this.bandWitdh - 0.5 * this.stoneSide}px`;
			stoneDiv.style.width = `${this.stoneSide}px`;
			stoneDiv.style.height = `${this.stoneSide}px`;
			stoneDiv.onclick = () => {
				alert("Location " + coordsToSgfCoords(row, column));

			}
			this.gobanDiv.appendChild(stoneDiv);
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
				centerDiv.style.color = blackStoneColor;
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

	animationTimeout: any;
	animationInterval: any;
	position = 0;

	public animate(initDelay?: number, interval?: number) {
		this.stopAnimation();
		this.drawBoard(0);
		let n = 0;
		this.animationTimeout = setTimeout(() => {
			this.drawBoard(++n);
			if (n >= this.positions.length - 1) {
				return;
			}
			this.animationInterval = setInterval(() => {
				this.drawBoard(++n);
			}, interval)
		}, initDelay);
	}

	public stopAnimation() {
		clearTimeout(this.animationTimeout);
		clearInterval(this.animationInterval);
	}

	public next() {
		this.stopAnimation();
		this.drawBoard(this.position + 1);
	}
	public previous() {
		this.stopAnimation();
		this.drawBoard(this.position - 1);
	}
	public first() {
		this.stopAnimation();
		this.drawBoard(0);
	}
	public last() {
		this.stopAnimation();
		this.drawBoard(this.positions.length - 1);
	}

	public async sgf() {
		let sgf = "(";
		for (let n = 0; n < this.positions.length; n++) {
			if (n > 0) {
				sgf += "\n;"
			}
			const pos = this.positions[n];
			for (const tag in pos.tags) {
				for (let val of pos.tags[tag]) {
					sgf += "\n" + `${tag}[${val}]`;
				}
			}
			for (let lineNo = 0; lineNo < pos.lines.length; lineNo++) {
				const line = pos.lines[lineNo].trim();
				console.log("line:", line);
				for (let columnNo = 0; columnNo < line.length; columnNo++) {
					const loc = line[columnNo];
					switch (loc) {
						case "w":
							if (n == 0) {
								sgf += "\n" + `AW[${this.toSgfCoordinates(lineNo, columnNo)}]`
							}
							break;
						case "W":
							if (!pos.tags["W"]) {
								sgf += "\n" + `W[${this.toSgfCoordinates(lineNo, columnNo)}]`
							}
							break;
						case "b":
							if (n == 0) {
								sgf += "\n" + `AB[${this.toSgfCoordinates(lineNo, columnNo)}]`
							}
							break;
						case "B":
							if (!pos.tags["B"]) {
								sgf += "\n" + `B[${this.toSgfCoordinates(lineNo, columnNo)}]`
							}
							break;
					}
				}
			}
		}
		sgf += "\n)";
		try {
			await navigator.clipboard.writeText(sgf);
			alert("Copied to clipboard");
		} catch (e) {
			let commentsEl = document.getElementById("goban_comment");
			commentsEl.innerHTML = "SGF:<br/>";
			const textarea = document.createElement("textarea")
			textarea.value = sgf;
			commentsEl.appendChild(textarea);
			textarea.select();
			/* Not working in anki:
			var element = document.createElement('a');
			element.setAttribute('href', 'data:application/x-go-sgf;charset=utf-8,' + encodeURIComponent(sgf));
			const fileName = new Date().toJSON().replace(/[^\d]/g, "") + ".sgf";
			element.setAttribute('download', fileName);
		  
			element.style.display = 'none';
			document.body.appendChild(element);
		  
			element.click();
		  
			document.body.removeChild(element);
			alert("Downloaded " + fileName);
			*/
		}
	}

	private toSgfCoordinates(lineNo: number, columnNo: number) {
		const coords = "abcdefghijklmnopqrs";
		return coords[columnNo] + coords[lineNo];
	}
}
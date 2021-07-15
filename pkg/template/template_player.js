var item = -1;

function call(o, m) {
	if (o && o[m]) o[m]()
}

function play(m) {
	player.hidden = false;
	call(player.children[0], "remove");
	player.append(m);
	m.controls = true;
	call(m, "play");
	m.focus()
};

const player = document.getElementById("player"),
	nf = [...document.getElementsByClassName("nf")].map(e => e.innerText),
	ls = [...document.querySelectorAll("button[data]")].map((b, id) => {
		const {
			a,
			i,
			v
		} = JSON.parse(atob(b.attributes.data.value)),
			p = a ? () => play(new Audio(a)) :
			i ? () => {
				const ii = new Image();
				ii.src = i;
				play(ii)
			} : v ? () => {
				const vv = document.createElement("video"),
					s = v.replace(/\.\w+$/, ".");
				vv.src = v;
				vv.poster = nf.find(n => n.startsWith(s) && /(?:bmp|jpeg|png|webp)$/.test(n)) || "";
				nf.filter(n => n.startsWith(s) && /\.(?:srt|vtt)$/.test(n)).forEach(t => {
					if (t.endsWith(".vtt")) {
						const tt = document.createElement("track");
						vv.append(tt);
						tt.kind = "subtitles";
						tt.language = tt.label = tt.src = t;
					} else {
						const tt = vv.addTextTrack("subtitles", t, t),
							p = t => t.replace(',', '.').split(':').reverse().reduce((s, f, i) => s += parseFloat(f) * (60 ** i || 1), 0);
						fetch(t).then(r => r.text()).then(tc => {
							for (const m of tc.matchAll(/([\d,:]+)\s+-->\s+([\d,:]+)\s+(.*)\n\r?\n/g))
								tt.addCue(new VTTCue(p(m[1]), p(m[2]), m[3]))
						});
					};
				});
				play(vv)
			} : () => {};
		b.addEventListener("click", () => {
			p();
			item = id
		});
		return p;
	});

document.addEventListener("keydown", e => {
	if (e.key == "Tab")
		ls[
			item = e.shiftKey ?
			(item <= 0 ? ls.length : item) - 1 :
			(item + 1) % ls.length
		](), e.preventDefault();
});
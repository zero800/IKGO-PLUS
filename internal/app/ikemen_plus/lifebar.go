package main

import (
	"fmt"
	"math"
	"strings"
)

type FinishType int32

const (
	FT_NotYet FinishType = iota
	FT_KO
	FT_DKO
	FT_TO
	FT_TODraw
)

type WinType int32

const (
	WT_N WinType = iota
	WT_S
	WT_H
	WT_C
	WT_T
	WT_Throw
	WT_Suicide
	WT_Teammate
	WT_Perfect
	WT_NumTypes
	WT_PN
	WT_PS
	WT_PH
	WT_PC
	WT_PT
	WT_PThrow
	WT_PSuicide
	WT_PTeammate
)

type RoundType int32

const (
	RT_Normal RoundType = iota
	RT_Deciding
	RT_Final
)

func (wt *WinType) SetPerfect() {
	if *wt >= WT_N && *wt <= WT_Teammate {
		*wt += WT_PN - WT_N
	}
}

type HealthBar struct {
	pos        [2]int32
	range_x    [2]int32
	bg0        AnimLayout
	bg1        AnimLayout
	bg2        AnimLayout
	mid        AnimLayout
	front      AnimLayout
	oldlife    float32
	midlife    float32
	midlifeMin float32
	mlifetime  int32
}

func readHealthBar(pre string, is IniSection,
	sff *Sff, at AnimationTable) *HealthBar {
	hb := &HealthBar{oldlife: 1, midlife: 1, midlifeMin: 1}
	is.ReadI32(pre+"pos", &hb.pos[0], &hb.pos[1])
	is.ReadI32(pre+"range.x", &hb.range_x[0], &hb.range_x[1])
	hb.bg0 = *ReadAnimLayout(pre+"bg0.", is, sff, at, 0)
	hb.bg1 = *ReadAnimLayout(pre+"bg1.", is, sff, at, 0)
	hb.bg2 = *ReadAnimLayout(pre+"bg2.", is, sff, at, 0)
	hb.mid = *ReadAnimLayout(pre+"mid.", is, sff, at, 0)
	hb.front = *ReadAnimLayout(pre+"front.", is, sff, at, 0)
	return hb
}
func (hb *HealthBar) step(life float32, gethit bool) {
	if len(hb.mid.anim.frames) > 0 && gethit {
		if hb.mlifetime < 30 {
			hb.mlifetime = 30
			hb.midlife, hb.midlifeMin = hb.oldlife, hb.oldlife
		}
	} else {
		if hb.mlifetime > 0 {
			hb.mlifetime--
		}
		if len(hb.mid.anim.frames) > 0 && hb.mlifetime <= 0 &&
			life < hb.midlifeMin {
			hb.midlifeMin += (life - hb.midlifeMin) *
				(1 / (12 - (life-hb.midlifeMin)*144))
		} else {
			hb.midlifeMin = life
		}
		if (len(hb.mid.anim.frames) == 0 || hb.mlifetime <= 0) &&
			hb.midlife > hb.midlifeMin {
			hb.midlife += (hb.midlifeMin - hb.midlife) / 8
		}
		hb.oldlife = life
	}
	mlmin := MaxF(hb.midlifeMin, life)
	if hb.midlife < mlmin {
		hb.midlife += (mlmin - hb.midlife) / 2
	}
	hb.bg0.Action()
	hb.bg1.Action()
	hb.bg2.Action()
	hb.mid.Action()
	hb.front.Action()
}
func (hb *HealthBar) reset() {
	hb.bg0.Reset()
	hb.bg1.Reset()
	hb.bg2.Reset()
	hb.mid.Reset()
	hb.front.Reset()
}
func (hb *HealthBar) bgDraw(layerno int16) {
	hb.bg0.Draw(float32(hb.pos[0]), float32(hb.pos[1]), layerno)
	hb.bg1.Draw(float32(hb.pos[0]), float32(hb.pos[1]), layerno)
	hb.bg2.Draw(float32(hb.pos[0]), float32(hb.pos[1]), layerno)
}
func (hb *HealthBar) draw(layerno int16, life float32) {
	width := func(life float32) (r [4]int32) {
		r = sys.scrrect
		if hb.range_x[0] < hb.range_x[1] {
			r[0] = int32((float32(hb.pos[0]+hb.range_x[0])+
				float32(sys.gameWidth-320)/2)*sys.widthScale + 0.5)
			r[2] = int32(float32(hb.range_x[1]-hb.range_x[0]+1)*life*
				sys.widthScale + 0.5)
		} else {
			r[2] = int32(float32(hb.range_x[0]-hb.range_x[1]+1)*life*
				sys.widthScale + 0.5)
			r[0] = int32((float32(hb.pos[0]+hb.range_x[0]+1)+
				float32(sys.gameWidth-320)/2)*sys.widthScale+0.5) - r[2]
		}
		return
	}
	if len(hb.mid.anim.frames) == 0 || life > hb.midlife {
		life = hb.midlife
	}
	lr, mr := width(life), width(hb.midlife)
	if hb.range_x[0] < hb.range_x[1] {
		mr[0] += lr[2]
	}
	mr[2] -= Min(mr[2], lr[2])
	hb.mid.lay.DrawAnim(&mr, float32(hb.pos[0]), float32(hb.pos[1]), 1,
		layerno, &hb.mid.anim)
	hb.front.lay.DrawAnim(&lr, float32(hb.pos[0]), float32(hb.pos[1]), 1,
		layerno, &hb.front.anim)
}

type PowerBar struct {
	snd          *Snd
	pos          [2]int32
	range_x      [2]int32
	partial_bar	 int32
	bg0          AnimLayout
	bg1          AnimLayout
	bg2          AnimLayout
	mid          AnimLayout
	front        AnimLayout
	fullfront    AnimLayout
	counter_font [3]int32
	counter_lay  Layout
	level_snd    [3][2]int32
	midpower     float32
	midpowerMin  float32
	prevLevel    int32
}

func newPowerBar(snd *Snd) (pb *PowerBar) {
	pb = &PowerBar{snd: snd, counter_font: [3]int32{-1},
		level_snd: [...][2]int32{{-1}, {-1}, {-1}}}
	return
}
func readPowerBar(pre string, is IniSection,
	sff *Sff, at AnimationTable, snd *Snd) *PowerBar {
	pb := newPowerBar(snd)
	is.ReadI32(pre+"pos", &pb.pos[0], &pb.pos[1])
	is.ReadI32(pre+"range.x", &pb.range_x[0], &pb.range_x[1])
	is.ReadI32(pre+"partial_bar", &pb.partial_bar)
	pb.bg0 = *ReadAnimLayout(pre+"bg0.", is, sff, at, 0)
	pb.bg1 = *ReadAnimLayout(pre+"bg1.", is, sff, at, 0)
	pb.bg2 = *ReadAnimLayout(pre+"bg2.", is, sff, at, 0)
	pb.mid = *ReadAnimLayout(pre+"mid.", is, sff, at, 0)
	pb.front = *ReadAnimLayout(pre+"front.", is, sff, at, 0)
	pb.fullfront = *ReadAnimLayout(pre+"fullfront.", is, sff, at, 0)
	is.ReadI32(pre+"counter.font", &pb.counter_font[0], &pb.counter_font[1],
		&pb.counter_font[2])
	pb.counter_lay = *ReadLayout(pre+"counter.", is, 0)
	for i := range pb.level_snd {
		if !is.ReadI32(fmt.Sprintf("%vlevel%v.snd", pre, i+1), &pb.level_snd[i][0],
			&pb.level_snd[i][1]) {
			is.ReadI32(fmt.Sprintf("level%v.snd", i+1), &pb.level_snd[i][0],
				&pb.level_snd[i][1])
		}
	}
	return pb
}
func (pb *PowerBar) step(power float32, level int32) {
	pb.midpower -= 1.0 / 144
	if power < pb.midpowerMin {
		pb.midpowerMin += (power - pb.midpowerMin) *
			(1 / (12 - (power-pb.midpowerMin)*144))
	} else {
		pb.midpowerMin = power
	}
	if pb.midpower < pb.midpowerMin {
		pb.midpower = pb.midpowerMin
	}
	if level > pb.prevLevel {
		i := Min(2, level-1)
		pb.snd.play(pb.level_snd[i])
	}
	pb.prevLevel = level
	pb.bg0.Action()
	pb.bg1.Action()
	pb.bg2.Action()
	pb.mid.Action()
	pb.front.Action()
	pb.fullfront.Action()
}
func (pb *PowerBar) reset() {
	pb.bg0.Reset()
	pb.bg1.Reset()
	pb.bg2.Reset()
	pb.mid.Reset()
	pb.front.Reset()
	pb.fullfront.Reset()
}
func (pb *PowerBar) bgDraw(layerno int16) {
	pb.bg0.Draw(float32(pb.pos[0]), float32(pb.pos[1]), layerno)
	pb.bg1.Draw(float32(pb.pos[0]), float32(pb.pos[1]), layerno)
	pb.bg2.Draw(float32(pb.pos[0]), float32(pb.pos[1]), layerno)
}
func (pb *PowerBar) draw(layerno int16, power float32,
	level int32, f []*Fnt, fullpower float32, fullpowermax float32) {
	width := func(power float32) (r [4]int32) {
		r = sys.scrrect
		if pb.range_x[0] < pb.range_x[1] {
			r[0] = int32((float32(pb.pos[0]+pb.range_x[0])+
				float32(sys.gameWidth-320)/2)*sys.widthScale + 0.5)
			r[2] = int32(float32(pb.range_x[1]-pb.range_x[0]+1)*power*
				sys.widthScale + 0.5)
		} else {
			r[2] = int32(float32(pb.range_x[0]-pb.range_x[1]+1)*power*
				sys.widthScale + 0.5)
			r[0] = int32((float32(pb.pos[0]+pb.range_x[0]+1)+
				float32(sys.gameWidth-320)/2)*sys.widthScale+0.5) - r[2]
		}
		return
	}
	pr, mr := width(power), width(pb.midpower)
	
	//calcule partial bar
	lvpower := power * (fullpowermax / 1000.0) - float32(level)	
    if pb.partial_bar > 0{ 
		pr = width(lvpower)
			if((fullpowermax / 1000.0) == float32(level)){
				pr = width(1.0);
		    }  		
	} 	 	
	
	if pb.range_x[0] < pb.range_x[1] {
		mr[0] += pr[2]
	}
	mr[2] -= Min(mr[2], pr[2])
	pb.mid.lay.DrawAnim(&mr, float32(pb.pos[0]), float32(pb.pos[1]), 1,
		layerno, &pb.mid.anim)
	pb.front.lay.DrawAnim(&pr, float32(pb.pos[0]), float32(pb.pos[1]), 1,
		layerno, &pb.front.anim)
	if fullpower == fullpowermax{
		pb.fullfront.lay.DrawAnim(&pr, float32(pb.pos[0]), float32(pb.pos[1]), 1,
		layerno, &pb.fullfront.anim)
	}	
	if pb.counter_font[0] >= 0 && int(pb.counter_font[0]) < len(f) {
		pb.counter_lay.DrawText(float32(pb.pos[0]), float32(pb.pos[1]), 1,
			layerno, fmt.Sprintf("%v", level),
			f[pb.counter_font[0]], pb.counter_font[1], pb.counter_font[2])
	}
}

type LifeBarFace struct {
	pos               [2]int32
	bg                AnimLayout
	face_spr          [2]int32
	face              *Sprite
	face_lay          Layout
	teammate_pos      [2]int32
	teammate_spacing  [2]int32
	teammate_bg       AnimLayout
	teammate_ko       AnimLayout
	teammate_face_spr [2]int32
	teammate_face     []*Sprite
	teammate_face_lay Layout
	numko             int32
}

func newLifeBarFace() *LifeBarFace {
	return &LifeBarFace{face_spr: [2]int32{-1}, teammate_face_spr: [2]int32{-1}}
}
func readLifeBarFace(pre string, is IniSection,
	sff *Sff, at AnimationTable) *LifeBarFace {
	f := newLifeBarFace()
	is.ReadI32(pre+"pos", &f.pos[0], &f.pos[1])
	f.bg = *ReadAnimLayout(pre+"bg.", is, sff, at, 0)
	is.ReadI32(pre+"face.spr", &f.face_spr[0], &f.face_spr[1])
	f.face_lay = *ReadLayout(pre+"face.", is, 0)
	is.ReadI32(pre+"teammate.pos", &f.teammate_pos[0], &f.teammate_pos[1])
	is.ReadI32(pre+"teammate.spacing", &f.teammate_spacing[0],
		&f.teammate_spacing[1])
	f.teammate_bg = *ReadAnimLayout(pre+"teammate.bg.", is, sff, at, 0)
	f.teammate_ko = *ReadAnimLayout(pre+"teammate.ko.", is, sff, at, 0)
	is.ReadI32(pre+"teammate.face.spr", &f.teammate_face_spr[0],
		&f.teammate_face_spr[1])
	f.teammate_face_lay = *ReadLayout(pre+"teammate.face.", is, 0)
	return f
}
func (f *LifeBarFace) step() {
	f.bg.Action()
	f.teammate_bg.Action()
	f.teammate_ko.Action()
}
func (f *LifeBarFace) reset() {
	f.bg.Reset()
	f.teammate_bg.Reset()
	f.teammate_ko.Reset()
}
func (f *LifeBarFace) bgDraw(layerno int16) {
	f.bg.Draw(float32(f.pos[0]), float32(f.pos[1]), layerno)
}
func (f *LifeBarFace) draw(layerno int16, fx *PalFX, superplayer bool) {
	ob := sys.brightness
	if superplayer {
		sys.brightness = 256
	}
	f.face_lay.DrawSprite(float32(f.pos[0]), float32(f.pos[1]), layerno,
		f.face, fx)
	sys.brightness = ob
	i := int32(len(f.teammate_face)) - 1
	x := float32(f.teammate_pos[0] + f.teammate_spacing[0]*(i-1))
	y := float32(f.teammate_pos[1] + f.teammate_spacing[1]*(i-1))
	for ; i >= 0; i-- {
		if i != f.numko {
			f.teammate_bg.Draw(x, y, layerno)
			f.teammate_face_lay.DrawSprite(x, y, layerno, f.teammate_face[i], nil)
			if i < f.numko {
				f.teammate_ko.Draw(x, y, layerno)
			}
			x -= float32(f.teammate_spacing[0])
			y -= float32(f.teammate_spacing[1])
		}
	}
}

type LifeBarName struct {
	pos       [2]int32
	name_font [3]int32
	name_lay  Layout
	bg        AnimLayout
}

func newLifeBarName() *LifeBarName {
	return &LifeBarName{name_font: [3]int32{-1}}
}
func readLifeBarName(pre string, is IniSection,
	sff *Sff, at AnimationTable) *LifeBarName {
	n := newLifeBarName()
	is.ReadI32(pre+"pos", &n.pos[0], &n.pos[1])
	is.ReadI32(pre+"name.font", &n.name_font[0], &n.name_font[1],
		&n.name_font[2])
	n.name_lay = *ReadLayout(pre+"name.", is, 0)
	n.bg = *ReadAnimLayout(pre+"bg.", is, sff, at, 0)
	return n
}
func (n *LifeBarName) step()  { n.bg.Action() }
func (n *LifeBarName) reset() { n.bg.Reset() }
func (n *LifeBarName) bgDraw(layerno int16) {
	n.bg.Draw(float32(n.pos[0]), float32(n.pos[1]), layerno)
}
func (n *LifeBarName) draw(layerno int16, f []*Fnt, name string) {
	if n.name_font[0] >= 0 && int(n.name_font[0]) < len(f) {
		n.name_lay.DrawText(float32(n.pos[0]), float32(n.pos[1]), 1, layerno, name,
			f[n.name_font[0]], n.name_font[1], n.name_font[2])
	}
}

type LifeBarWinIcon struct {
	pos           [2]int32
	iconoffset    [2]int32
	useiconupto   int32
	counter_font  [3]int32
	counter_lay   Layout
	icon          [WT_NumTypes]AnimLayout
	wins          []WinType
	numWins       int
	added, addedP *Animation
}

func newLifeBarWinIcon() *LifeBarWinIcon {
	return &LifeBarWinIcon{useiconupto: 4, counter_font: [3]int32{-1}}
}
func readLifeBarWinIcon(pre string, is IniSection,
	sff *Sff, at AnimationTable) *LifeBarWinIcon {
	wi := newLifeBarWinIcon()
	is.ReadI32(pre+"pos", &wi.pos[0], &wi.pos[1])
	is.ReadI32(pre+"iconoffset", &wi.iconoffset[0], &wi.iconoffset[1])
	is.ReadI32("useiconupto", &wi.useiconupto)
	is.ReadI32(pre+"counter.font", &wi.counter_font[0], &wi.counter_font[1],
		&wi.counter_font[2])
	wi.counter_lay = *ReadLayout(pre+"counter.", is, 0)
	wi.icon[WT_N] = *ReadAnimLayout(pre+"n.", is, sff, at, 0)
	wi.icon[WT_S] = *ReadAnimLayout(pre+"s.", is, sff, at, 0)
	wi.icon[WT_H] = *ReadAnimLayout(pre+"h.", is, sff, at, 0)
	wi.icon[WT_C] = *ReadAnimLayout(pre+"c.", is, sff, at, 0)
	wi.icon[WT_T] = *ReadAnimLayout(pre+"t.", is, sff, at, 0)
	wi.icon[WT_Throw] = *ReadAnimLayout(pre+"throw.", is, sff, at, 0)
	wi.icon[WT_Suicide] = *ReadAnimLayout(pre+"suicide.", is, sff, at, 0)
	wi.icon[WT_Teammate] = *ReadAnimLayout(pre+"teammate.", is, sff, at, 0)
	wi.icon[WT_Perfect] = *ReadAnimLayout(pre+"perfect.", is, sff, at, 0)
	return wi
}
func (wi *LifeBarWinIcon) add(wt WinType) {
	wi.wins = append(wi.wins, wt)
	if wt >= WT_PN {
		wi.addedP = &Animation{}
		*wi.addedP = wi.icon[WT_Perfect].anim
		wi.addedP.Reset()
		wt -= WT_PN
	}
	wi.added = &Animation{}
	*wi.added = wi.icon[wt].anim
	wi.added.Reset()
}
func (wi *LifeBarWinIcon) step(numwin int32) {
	if int(numwin) < len(wi.wins) {
		wi.wins = wi.wins[:numwin]
		wi.reset()
	}
	for i := range wi.icon {
		wi.icon[i].Action()
	}
	if wi.added != nil {
		wi.added.Action()
	}
	if wi.addedP != nil {
		wi.addedP.Action()
	}
}
func (wi *LifeBarWinIcon) reset() {
	for i := range wi.icon {
		wi.icon[i].Reset()
	}
	wi.numWins = len(wi.wins)
	wi.added, wi.addedP = nil, nil
}
func (wi *LifeBarWinIcon) clear() { wi.wins = nil }
func (wi *LifeBarWinIcon) draw(layerno int16, f []*Fnt) {
	if len(wi.wins) > int(wi.useiconupto) {
		if wi.counter_font[0] >= 0 && int(wi.counter_font[0]) < len(f) {
			wi.counter_lay.DrawText(float32(wi.pos[0]), float32(wi.pos[1]), 1,
				layerno, fmt.Sprintf("%v", len(wi.wins)),
				f[wi.counter_font[0]], wi.counter_font[1], wi.counter_font[2])
		}
	} else {
		i := 0
		for ; i < wi.numWins; i++ {
			wt, p := wi.wins[i], false
			if wt >= WT_PN {
				wt -= WT_PN
				p = true
			}
			wi.icon[wt].Draw(float32(wi.pos[0]+wi.iconoffset[0]*int32(i)),
				float32(wi.pos[1]+wi.iconoffset[1]*int32(i)), layerno)
			if p {
				wi.icon[WT_Perfect].Draw(float32(wi.pos[0]+wi.iconoffset[0]*int32(i)),
					float32(wi.pos[1]+wi.iconoffset[1]*int32(i)), layerno)
			}
		}
		if wi.added != nil {
			wt, p := wi.wins[i], false
			if wi.addedP != nil {
				wt -= WT_PN
				p = true
			}
			wi.icon[wt].lay.DrawAnim(&sys.scrrect,
				float32(wi.pos[0]+wi.iconoffset[0]*int32(i)),
				float32(wi.pos[1]+wi.iconoffset[1]*int32(i)), 1, layerno, wi.added)
			if p {
				wi.icon[WT_Perfect].lay.DrawAnim(&sys.scrrect,
					float32(wi.pos[0]+wi.iconoffset[0]*int32(i)),
					float32(wi.pos[1]+wi.iconoffset[1]*int32(i)), 1, layerno, wi.addedP)
			}
		}
	}
}

type LifeBarTime struct {
	pos            [2]int32
	counter_font   [3]int32
	counter_lay    Layout
	bg             AnimLayout
	framespercount int32
}

func newLifeBarTime() *LifeBarTime {
	return &LifeBarTime{counter_font: [3]int32{-1}, framespercount: 60}
}
func readLifeBarTime(is IniSection,
	sff *Sff, at AnimationTable) *LifeBarTime {
	t := newLifeBarTime()
	is.ReadI32("pos", &t.pos[0], &t.pos[1])
	is.ReadI32("counter.font", &t.counter_font[0], &t.counter_font[1],
		&t.counter_font[2])
	t.counter_lay = *ReadLayout("counter.", is, 0)
	t.bg = *ReadAnimLayout("bg.", is, sff, at, 0)
	is.ReadI32("framespercount", &t.framespercount)
	return t
}
func (t *LifeBarTime) step()  { t.bg.Action() }
func (t *LifeBarTime) reset() { t.bg.Reset() }
func (t *LifeBarTime) bgDraw(layerno int16) {
	t.bg.Draw(float32(t.pos[0]), float32(t.pos[1]), layerno)
}
func (t *LifeBarTime) draw(layerno int16, f []*Fnt) {
	if t.framespercount > 0 &&
		t.counter_font[0] >= 0 && int(t.counter_font[0]) < len(f) {
		time := "o"
		if sys.time >= 0 {
			time = fmt.Sprintf("%v", sys.time/t.framespercount)
		}
		t.counter_lay.DrawText(float32(t.pos[0]), float32(t.pos[1]), 1, layerno,
			time, f[t.counter_font[0]], t.counter_font[1], t.counter_font[2])
	}
}

type LifeBarCombo struct {
	pos           [2]int32
	start_x       float32
	counter_font  [3]int32
	counter_shake bool
	counter_lay   Layout
	text_font     [3]int32
	text_text     string
	text_lay      Layout
	displaytime   int32
	cur, old      [2]int32
	resttime      [2]int32
	counterX      [2]float32
	shaketime     [2]int32
}

func newLifeBarCombo() *LifeBarCombo {
	return &LifeBarCombo{counter_font: [3]int32{-1}, text_font: [3]int32{-1},
		displaytime: 90}
}
func readLifeBarCombo(is IniSection) *LifeBarCombo {
	c := newLifeBarCombo()
	is.ReadI32("pos", &c.pos[0], &c.pos[1])
	is.ReadF32("start.x", &c.start_x)
	is.ReadI32("counter.font", &c.counter_font[0], &c.counter_font[1],
		&c.counter_font[2])
	is.ReadBool("counter.shake", &c.counter_shake)
	c.counter_lay = *ReadLayout("counter.", is, 2)
	c.counter_lay.offset = [2]float32{}
	is.ReadI32("text.font", &c.text_font[0], &c.text_font[1], &c.text_font[2])
	c.text_text = is["text.text"]
	c.text_lay = *ReadLayout("text.", is, 2)
	is.ReadI32("displaytime", &c.displaytime)
	return c
}
func (c *LifeBarCombo) step(combo [2]int32) {
	for i := range c.cur {
		if c.resttime[i] > 0 {
			c.counterX[i] -= c.counterX[i] / 8
		} else {
			c.counterX[i] -= sys.lifebarFontScale * 4
			if c.counterX[i] < c.start_x*2 {
				c.counterX[i] = c.start_x * 2
			}
		}
		if c.shaketime[i] > 0 {
			c.shaketime[i]--
		}
		if AbsF(c.counterX[i]) < 1 {
			c.resttime[i]--
		}
		if combo[i] >= 2 && c.old[i] != combo[i] {
			c.cur[i] = combo[i]
			c.resttime[i] = c.displaytime
			if c.counter_shake {
				c.shaketime[i] = 15
			}
		}
		c.old[i] = combo[i]
	}
}
func (c *LifeBarCombo) reset() {
	c.cur, c.old, c.resttime = [2]int32{}, [2]int32{}, [2]int32{}
	c.counterX = [...]float32{c.start_x * 2, c.start_x * 2}
	c.shaketime = [2]int32{}
}
func (c *LifeBarCombo) draw(layerno int16, f []*Fnt) {
	haba := func(n int32) float32 {
		if c.counter_font[0] < 0 || int(c.counter_font[0]) >= len(f) {
			return 0
		}
		return float32(f[c.counter_font[0]].TextWidth(fmt.Sprintf("%v", n)))
	}
	for i := range c.cur {
		if c.resttime[i] <= 0 && c.counterX[i] == c.start_x*2 {
			continue
		}
		var x float32
		if i&1 == 0 {
			if c.start_x <= 0 {
				x = c.counterX[i]
			}
			x += float32(c.pos[0]) + haba(c.cur[i])
		} else {
			if c.start_x <= 0 {
				x = -c.counterX[i]
			}
			x += 320 - float32(c.pos[0])
		}
		if c.text_font[0] >= 0 && int(c.text_font[0]) < len(f) {
			text := OldSprintf(c.text_text, c.cur[i])
			if i&1 == 0 {
				if c.pos[0] != 0 {
					x += c.text_lay.offset[0] *
						((1 - sys.lifebarFontScale) * sys.lifebarFontScale)
				}
			} else {
				tmp := c.text_lay.offset[0]
				if c.pos[0] == 0 {
					tmp *= sys.lifebarFontScale
				}
				x -= tmp + float32(f[c.text_font[0]].TextWidth(text))*
					c.text_lay.scale[0]*sys.lifebarFontScale
			}
			c.text_lay.DrawText(x, float32(c.pos[1]), 1, layerno,
				text, f[c.text_font[0]], c.text_font[1], 1)
		}
		if c.counter_font[0] >= 0 && int(c.counter_font[0]) < len(f) {
			z := 1 + float32(c.shaketime[i])*(1.0/20)*
				float32(math.Sin(float64(c.shaketime[i])*(math.Pi/2.5)))
			c.counter_lay.DrawText(x/z, float32(c.pos[1])/z, z, layerno,
				fmt.Sprintf("%v", c.cur[i]), f[c.counter_font[0]],
				c.counter_font[1], -1)
		}
	}
}

type LifeBarRound struct {
	snd                *Snd
	pos                [2]int32
	match_wins         int32
	match_maxdrawgames int32
	start_waittime     int32
	round_time         int32
	round_sndtime      int32
	round_default      AnimTextSnd
	round              [9]AnimTextSnd
	fight_time         int32
	fight_sndtime      int32
	fight              AnimTextSnd
	ctrl_time          int32
	ko_time            int32
	ko_sndtime         int32
	ko, dko, to        AnimTextSnd
	n, s, h, throw, c  AnimTextSnd
	t, suicide         AnimTextSnd
	teammate, perfect  AnimTextSnd
	slow_time          int32
	over_waittime      int32
	over_hittime       int32
	over_wintime       int32
	over_time          int32
	win_time           int32
	win_sndtime        int32
	win, win2, drawn   AnimTextSnd
	cur                int32
	wt, swt, dt        [2]int32
	fnt                []*Fnt
	timerActive        bool
}

func newLifeBarRound(snd *Snd, fnt []*Fnt) *LifeBarRound {
	return &LifeBarRound{snd: snd, match_wins: 2, match_maxdrawgames: 1,
		start_waittime: 30, ctrl_time: 30, slow_time: 60, over_waittime: 45,
		over_hittime: 10, over_wintime: 45, over_time: 210, win_sndtime: 60,
		fnt: fnt}
}
func readLifeBarRound(is IniSection,
	sff *Sff, at AnimationTable, snd *Snd, fnt []*Fnt) *LifeBarRound {
	r := newLifeBarRound(snd, fnt)
	var tmp int32
	is.ReadI32("pos", &r.pos[0], &r.pos[1])
	tmp = Atoi(sys.cmdFlags["-rounds"])
	if tmp > 0 {
		r.match_wins = tmp
	} else {
		is.ReadI32("match.wins", &r.match_wins)
	}
	is.ReadI32("match.maxdrawgames", &r.match_maxdrawgames)
	if is.ReadI32("start.waittime", &tmp) {
		r.start_waittime = Max(1, tmp)
	}
	is.ReadI32("round.time", &r.round_time)
	is.ReadI32("round.sndtime", &r.round_sndtime)
	r.round_default = *ReadAnimTextSnd("round.default.", is, sff, at, 2)
	for i := range r.round {
		r.round[i] = r.round_default
		r.round[i].Read(fmt.Sprintf("round%v.", i+1), is, at, 2)
	}
	is.ReadI32("fight.time", &r.fight_time)
	is.ReadI32("fight.sndtime", &r.fight_sndtime)
	r.fight = *ReadAnimTextSnd("fight.", is, sff, at, 2)
	if is.ReadI32("ctrl.time", &tmp) {
		r.ctrl_time = Max(1, tmp)
	}
	is.ReadI32("ko.time", &r.ko_time)
	is.ReadI32("ko.sndtime", &r.ko_sndtime)
	r.ko = *ReadAnimTextSnd("ko.", is, sff, at, 1)
	r.dko = *ReadAnimTextSnd("dko.", is, sff, at, 1)
	r.to = *ReadAnimTextSnd("to.", is, sff, at, 1)
	r.n = *ReadAnimTextSnd("n.", is, sff, at, 1)
	r.s = *ReadAnimTextSnd("s.", is, sff, at, 1)
	r.h = *ReadAnimTextSnd("h.", is, sff, at, 1)
	r.throw = *ReadAnimTextSnd("throw.", is, sff, at, 1)
	r.c = *ReadAnimTextSnd("c.", is, sff, at, 1)
	r.t = *ReadAnimTextSnd("t.", is, sff, at, 1)
	r.suicide = *ReadAnimTextSnd("suicide.", is, sff, at, 1)
	r.teammate = *ReadAnimTextSnd("teammate.", is, sff, at, 1)
	r.perfect = *ReadAnimTextSnd("perfect.", is, sff, at, 1)
	is.ReadI32("slow.time", &r.slow_time)
	if is.ReadI32("over.hittime", &tmp) {
		r.over_hittime = Max(0, tmp)
	}
	if is.ReadI32("over.waittime", &tmp) {
		r.over_waittime = Max(1, tmp)
	}
	if is.ReadI32("over.wintime", &tmp) {
		r.over_wintime = Max(1, tmp)
	}
	if is.ReadI32("over.time", &tmp) {
		r.over_time = Max(r.over_wintime+1, tmp)
	}
	is.ReadI32("win.time", &r.win_time)
	is.ReadI32("win.sndtime", &r.win_sndtime)
	r.win = *ReadAnimTextSnd("win.", is, sff, at, 1)
	r.win2 = *ReadAnimTextSnd("win2.", is, sff, at, 1)
	r.drawn = *ReadAnimTextSnd("draw.", is, sff, at, 1)
	return r
}
func (r *LifeBarRound) callFight() {
	r.fight.Reset()
	r.cur, r.wt[0], r.swt[0], r.dt[0] = 1, r.fight_time, r.fight_sndtime, 0
	sys.timerCount = append(sys.timerCount, sys.gameTime)
	r.timerActive = true
}
func (r *LifeBarRound) act() bool {
	if sys.intro > r.ctrl_time {
		r.cur, r.wt[0], r.swt[0], r.dt[0] = 0, r.round_time, r.round_sndtime, 0
	} else if sys.intro >= 0 || r.cur < 2 {
		if !sys.tickNextFrame() {
			return false
		}
		switch r.cur {
		case 0:
			if r.swt[0] == 0 {
				if int(sys.round) <= len(r.round) {
					r.snd.play(r.round[sys.round-1].snd)
				} else {
					r.snd.play(r.round_default.snd)
				}
			}
			r.swt[0]--
			if r.wt[0] <= 0 {
				r.dt[0]++
				end := false
				if int(sys.round) <= len(r.round) {
					r.round[sys.round-1].Action()
					end = r.round[sys.round-1].End(r.dt[0])
				} else {
					r.round_default.Action()
					end = r.round_default.End(r.dt[0])
				}
				if end {
					r.callFight()
					return true
				}
			}
			r.wt[0]--
			return false
		case 1:
			if r.swt[0] == 0 {
				r.snd.play(r.fight.snd)
			}
			r.swt[0]--
			if r.wt[0] <= 0 {
				r.dt[0]++
				r.fight.Action()
				if r.fight.End(r.dt[0]) {
					r.cur, r.wt[0], r.swt[0], r.dt[0] = 2, r.ko_time, r.ko_sndtime, 0
					r.wt[1], r.swt[1], r.dt[1] = r.win_time, r.win_sndtime, 0
					break
				}
			}
			r.wt[0]--
		}
	} else if r.cur == 2 && (sys.finish != FT_NotYet || sys.time == 0) {
		if r.timerActive {
			if sys.gameTime-sys.timerCount[sys.round-1] > 0 {
				sys.timerCount[sys.round-1] = sys.gameTime - sys.timerCount[sys.round-1]
			} else {
				sys.timerCount[sys.round-1] = 0
			}
			r.timerActive = false
		}
		f := func(ats *AnimTextSnd, t int) {
			if r.swt[t] == 0 {
				r.snd.play(ats.snd)
			}
			r.swt[t]--
			if ats.End(r.dt[t]) {
				r.wt[t] = 2
			}
			if r.wt[t] <= 0 {
				r.dt[t]++
				ats.Action()
			}
			r.wt[t]--
		}
		switch sys.finish {
		case FT_KO:
			f(&r.ko, 0)
		case FT_DKO:
			f(&r.dko, 0)
		default:
			f(&r.to, 0)
		}
		if sys.winTeam >= 0 {
			switch sys.winType[sys.winTeam] {
			case WT_N, WT_PN:
				f(&r.n, 0)
			case WT_S, WT_PS:
				f(&r.s, 0)
			case WT_H, WT_PH:
				f(&r.h, 0)
			case WT_C, WT_PC:
				f(&r.c, 0)
			case WT_T, WT_PT:
				f(&r.t, 0)
			case WT_Throw, WT_PThrow:
				f(&r.throw, 0)
			case WT_Suicide, WT_PSuicide:
				f(&r.suicide, 0)
			case WT_Teammate, WT_PTeammate:
				f(&r.teammate, 0)
			}
			if sys.winType[sys.winTeam] >= WT_Perfect {
				f(&r.perfect, 0)
			}
		}
		if sys.intro < -(r.over_hittime + r.over_waittime + r.over_wintime) {
			if sys.finish == FT_DKO || sys.finish == FT_TODraw {
				f(&r.drawn, 1)
			} else if sys.winTeam >= 0 && sys.tmode[sys.winTeam] == TM_Simul {
				f(&r.win2, 1)
			} else {
				f(&r.win, 1)
			}
		}
	}
	return sys.tickNextFrame()
}
func (r *LifeBarRound) reset() {
	r.round_default.Reset()
	for i := range r.round {
		r.round[i].Reset()
	}
	r.fight.Reset()
	r.ko.Reset()
	r.dko.Reset()
	r.to.Reset()
	r.n.Reset()
	r.s.Reset()
	r.h.Reset()
	r.throw.Reset()
	r.c.Reset()
	r.t.Reset()
	r.suicide.Reset()
	r.teammate.Reset()
	r.perfect.Reset()
	r.win.Reset()
	r.win2.Reset()
	r.drawn.Reset()
}
func (r *LifeBarRound) draw(layerno int16) {
	ob := sys.brightness
	sys.brightness = 255
	switch r.cur {
	case 0:
		if r.wt[0] < 0 && sys.intro <= r.ctrl_time {
			if int(sys.round) <= len(r.round) {
				tmp := r.round[sys.round-1].text
				r.round[sys.round-1].text = OldSprintf(tmp, sys.round)
				r.round[sys.round-1].Draw(float32(r.pos[0]), float32(r.pos[1]),
					layerno, r.fnt)
				r.round[sys.round-1].text = tmp
			} else {
				tmp := r.round_default.text
				r.round_default.text = OldSprintf(tmp, sys.round)
				r.round_default.Draw(float32(r.pos[0]), float32(r.pos[1]),
					layerno, r.fnt)
				r.round_default.text = tmp
			}
		}
	case 1:
		if r.wt[0] < 0 {
			r.fight.Draw(float32(r.pos[0]), float32(r.pos[1]), layerno, r.fnt)
		}
	case 2:
		if r.wt[0] < 0 {
			switch sys.finish {
			case FT_KO:
				r.ko.Draw(float32(r.pos[0]), float32(r.pos[1]), layerno, r.fnt)
			case FT_DKO:
				r.dko.Draw(float32(r.pos[0]), float32(r.pos[1]), layerno, r.fnt)
			default:
				r.to.Draw(float32(r.pos[0]), float32(r.pos[1]), layerno, r.fnt)
			}
		}
		if r.wt[1] < 0 {
			if sys.finish == FT_DKO || sys.finish == FT_TODraw {
				r.drawn.Draw(float32(r.pos[0]), float32(r.pos[1]), layerno, r.fnt)
			} else if sys.tmode[sys.winTeam] == TM_Simul {
				tmp := r.win2.text
				var inter []interface{}
				for i := sys.winTeam; i < len(sys.chars); i += 2 {
					if len(sys.chars[i]) > 0 {
						inter = append(inter, sys.cgi[i].displayname)
					}
				}
				r.win2.text = OldSprintf(tmp, inter...)
				r.win2.Draw(float32(r.pos[0]), float32(r.pos[1]), layerno, r.fnt)
				r.win2.text = tmp
			} else {
				tmp := r.win.text
				r.win.text = OldSprintf(tmp, sys.cgi[sys.winTeam].displayname)
				r.win.Draw(float32(r.pos[0]), float32(r.pos[1]), layerno, r.fnt)
				r.win.text = tmp
			}
		}
		if (sys.finish != FT_NotYet || sys.time == 0) && sys.winTeam >= 0 {
			switch sys.winType[sys.winTeam] {
			case WT_N, WT_PN:
				r.n.Draw(float32(r.pos[0]), float32(r.pos[1]), layerno, r.fnt)
			case WT_S, WT_PS:
				r.s.Draw(float32(r.pos[0]), float32(r.pos[1]), layerno, r.fnt)
			case WT_H, WT_PH:
				r.h.Draw(float32(r.pos[0]), float32(r.pos[1]), layerno, r.fnt)
			case WT_C, WT_PC:
				r.c.Draw(float32(r.pos[0]), float32(r.pos[1]), layerno, r.fnt)
			case WT_T, WT_PT:
				r.t.Draw(float32(r.pos[0]), float32(r.pos[1]), layerno, r.fnt)
			case WT_Throw, WT_PThrow:
				r.throw.Draw(float32(r.pos[0]), float32(r.pos[1]), layerno, r.fnt)
			case WT_Suicide, WT_PSuicide:
				r.suicide.Draw(float32(r.pos[0]), float32(r.pos[1]), layerno, r.fnt)
			case WT_Teammate, WT_PTeammate:
				r.teammate.Draw(float32(r.pos[0]), float32(r.pos[1]), layerno, r.fnt)
			}
			if sys.winType[sys.winTeam] >= WT_Perfect {
				r.perfect.Draw(float32(r.pos[0]), float32(r.pos[1]), layerno, r.fnt)
			}
		}
	}
	sys.brightness = ob
}

type LifeBarChallenger struct {
	cnt        int32
	snd        *Snd
	pos        [2]int32
	challenger AnimTextSnd
	sndtime    int32
	over_pause int32
	over_time  int32
	bg         AnimLayout
	fnt        []*Fnt
}

func newLifeBarChallenger(snd *Snd, fnt []*Fnt) *LifeBarChallenger {
	return &LifeBarChallenger{snd: snd, fnt: fnt}
}
func readLifeBarChallenger(is IniSection,
	sff *Sff, at AnimationTable, snd *Snd, fnt []*Fnt) *LifeBarChallenger {
	ch := newLifeBarChallenger(snd, fnt)
	is.ReadI32("pos", &ch.pos[0], &ch.pos[1])
	ch.challenger = *ReadAnimTextSnd("", is, sff, at, 1)
	is.ReadI32("sndtime", &ch.sndtime)
	var tmp int32
	if is.ReadI32("over.pause", &tmp) {
		ch.over_pause = Max(1, tmp)
	}
	if is.ReadI32("over.time", &tmp) {
		ch.over_time = Max(ch.over_pause+1, tmp)
	}
	ch.bg = *ReadAnimLayout("bg.", is, sff, at, 0)
	return ch
}
func (ch *LifeBarChallenger) step() {
	if sys.challenger > 0 {
		if ch.cnt == ch.sndtime {
			ch.snd.play(ch.challenger.snd)
		}
		if ch.cnt == ch.over_pause {
			sys.paused = true
		}
		if ch.challenger.displaytime > ch.cnt {
			ch.challenger.Action()
			ch.bg.Action()
		}
		ch.cnt += 1
	}
}
func (ch *LifeBarChallenger) reset() {
	ch.cnt = 0
	ch.challenger.Reset()
	ch.bg.Reset()
}
func (ch *LifeBarChallenger) bgDraw(layerno int16) {
	ch.bg.Draw(float32(ch.pos[0]), float32(ch.pos[1]), layerno)
}
func (ch *LifeBarChallenger) draw(layerno int16, f []*Fnt) {
	ch.challenger.Draw(float32(ch.pos[0]), float32(ch.pos[1]), layerno, ch.fnt)
}

type Lifebar struct {
	fat       AnimationTable
	fsff      *Sff
	snd, fsnd *Snd
	fnt       [10]*Fnt
	ref       [4][2]int
	num       [4][2]int
	hb        [8][]*HealthBar
	pb        [6][]*PowerBar
	fa        [4][]*LifeBarFace
	nm        [4][]*LifeBarName
	wi        [2]*LifeBarWinIcon
	ti        *LifeBarTime
	co        *LifeBarCombo
	ro        *LifeBarRound
	ch        *LifeBarChallenger
	bgdef     *BGDef
}

func loadLifebar(deffile string) (*Lifebar, error) {
	str, err := LoadText(deffile)
	if err != nil {
		return nil, err
	}
	l := &Lifebar{fsff: &Sff{}, snd: &Snd{},
		hb: [...][]*HealthBar{make([]*HealthBar, 2), make([]*HealthBar, 4),
			make([]*HealthBar, 2), make([]*HealthBar, 4), make([]*HealthBar, 6),
			make([]*HealthBar, 8), make([]*HealthBar, 6), make([]*HealthBar, 8)},
		pb: [...][]*PowerBar{make([]*PowerBar, 2), make([]*PowerBar, 4),
			make([]*PowerBar, 2), make([]*PowerBar, 2), make([]*PowerBar, 6),
			make([]*PowerBar, 8)},
		fa: [...][]*LifeBarFace{make([]*LifeBarFace, 2), make([]*LifeBarFace, 8),
			make([]*LifeBarFace, 2), make([]*LifeBarFace, 8)},
		nm: [...][]*LifeBarName{make([]*LifeBarName, 2), make([]*LifeBarName, 8),
			make([]*LifeBarName, 2), make([]*LifeBarName, 8)}}
	missing := map[string]int{"[simul_3p lifebar]": 3, "[simul_4p lifebar]": 4,
		"[tag lifebar]": 5, "[tag_3p lifebar]": 6, "[tag_4p lifebar]": 7,
		"[simul powerbar]": 1, "[turns powerbar]": 2, "[simul_3p powerbar]": 3,
		"[simul_4p powerbar]": 4, "[tag powerbar]": 5, "[tag face]": -1,
		"[tag name]": -1, "[challenger]": -1}
	strc := strings.ToLower(strings.TrimSpace(str))
	for k := range missing {
		strc = strings.Replace(strc, ";"+k, "", -1)
		if strings.Contains(strc, k) {
			delete(missing, k)
		} else {
			str += "\n" + k
		}
	}
	sff, lines, i := &Sff{}, SplitAndTrim(str, "\n"), 0
	at := ReadAnimationTable(sff, lines, &i)
	i = 0
	filesflg := true
	for i < len(lines) {
		is, name, subname := ReadIniSection(lines, &i)
		switch name {
		case "files":
			if filesflg {
				filesflg = false
				if is.LoadFile("sff", deffile, func(filename string) error {
					s, err := loadSff(filename, false)
					if err != nil {
						return err
					}
					*sff = *s
					return nil
				}); err != nil {
					return nil, err
				}
				if is.LoadFile("snd", deffile, func(filename string) error {
					s, err := LoadSnd(filename)
					if err != nil {
						return err
					}
					*l.snd = *s
					return nil
				}); err != nil {
					return nil, err
				}
				if is.LoadFile("fightfx.sff", deffile, func(filename string) error {
					s, err := loadSff(filename, false)
					if err != nil {
						return err
					}
					*l.fsff = *s
					return nil
				}); err != nil {
					return nil, err
				}
				if is.LoadFile("fightfx.air", deffile, func(filename string) error {
					str, err := LoadText(filename)
					if err != nil {
						return err
					}
					lines, i := SplitAndTrim(str, "\n"), 0
					l.fat = ReadAnimationTable(l.fsff, lines, &i)
					return nil
				}); err != nil {
					return nil, err
				}
				if is.LoadFile("common.snd", deffile, func(filename string) error {
					l.fsnd, err = LoadSnd(filename)
					return err
				}); err != nil {
					return nil, err
				}
				for i := range l.fnt {
					if is.LoadFile(fmt.Sprintf("font%v", i), deffile,
						func(filename string) error {
							h := int32(0)
							if len(is[fmt.Sprintf("font%v.height", i)]) > 0 {
								h = Atoi(is[fmt.Sprintf("font%v.height", i)])
							}
							l.fnt[i], err = loadFnt(filename, h)
							return err
						}); err != nil {
						return nil, err
					}
				}
			}
		case "fonts":
			is.ReadF32("scale", &sys.lifebarFontScale)
		case "lifebar":
			if l.hb[0][0] == nil {
				l.hb[0][0] = readHealthBar("p1.", is, sff, at)
			}
			if l.hb[0][1] == nil {
				l.hb[0][1] = readHealthBar("p2.", is, sff, at)
			}
		case "powerbar":
			if l.pb[0][0] == nil {
				l.pb[0][0] = readPowerBar("p1.", is, sff, at, l.snd)
			}
			if l.pb[0][1] == nil {
				l.pb[0][1] = readPowerBar("p2.", is, sff, at, l.snd)
			}
		case "face":
			if l.fa[0][0] == nil {
				l.fa[0][0] = readLifeBarFace("p1.", is, sff, at)
			}
			if l.fa[0][1] == nil {
				l.fa[0][1] = readLifeBarFace("p2.", is, sff, at)
			}
		case "name":
			if l.nm[0][0] == nil {
				l.nm[0][0] = readLifeBarName("p1.", is, sff, at)
			}
			if l.nm[0][1] == nil {
				l.nm[0][1] = readLifeBarName("p2.", is, sff, at)
			}
		case "simul ":
			subname = strings.ToLower(subname)
			switch {
			case len(subname) >= 7 && subname[:7] == "lifebar":
				if l.hb[1][0] == nil {
					l.hb[1][0] = readHealthBar("p1.", is, sff, at)
				}
				if l.hb[1][1] == nil {
					l.hb[1][1] = readHealthBar("p2.", is, sff, at)
				}
				if l.hb[1][2] == nil {
					l.hb[1][2] = readHealthBar("p3.", is, sff, at)
				}
				if l.hb[1][3] == nil {
					l.hb[1][3] = readHealthBar("p4.", is, sff, at)
				}
			case len(subname) >= 8 && subname[:8] == "powerbar":
				if l.pb[1][0] == nil {
					l.pb[1][0] = readPowerBar("p1.", is, sff, at, l.snd)
				}
				if l.pb[1][1] == nil {
					l.pb[1][1] = readPowerBar("p2.", is, sff, at, l.snd)
				}
				if l.pb[1][2] == nil {
					l.pb[1][2] = readPowerBar("p3.", is, sff, at, l.snd)
				}
				if l.pb[1][3] == nil {
					l.pb[1][3] = readPowerBar("p4.", is, sff, at, l.snd)
				}
			case len(subname) >= 4 && subname[:4] == "face":
				if l.fa[1][0] == nil {
					l.fa[1][0] = readLifeBarFace("p1.", is, sff, at)
				}
				if l.fa[1][1] == nil {
					l.fa[1][1] = readLifeBarFace("p2.", is, sff, at)
				}
				if l.fa[1][2] == nil {
					l.fa[1][2] = readLifeBarFace("p3.", is, sff, at)
				}
				if l.fa[1][3] == nil {
					l.fa[1][3] = readLifeBarFace("p4.", is, sff, at)
				}
				if l.fa[1][4] == nil {
					l.fa[1][4] = readLifeBarFace("p5.", is, sff, at)
				}
				if l.fa[1][5] == nil {
					l.fa[1][5] = readLifeBarFace("p6.", is, sff, at)
				}
				if l.fa[1][6] == nil {
					l.fa[1][6] = readLifeBarFace("p7.", is, sff, at)
				}
				if l.fa[1][7] == nil {
					l.fa[1][7] = readLifeBarFace("p8.", is, sff, at)
				}
			case len(subname) >= 4 && subname[:4] == "name":
				if l.nm[1][0] == nil {
					l.nm[1][0] = readLifeBarName("p1.", is, sff, at)
				}
				if l.nm[1][1] == nil {
					l.nm[1][1] = readLifeBarName("p2.", is, sff, at)
				}
				if l.nm[1][2] == nil {
					l.nm[1][2] = readLifeBarName("p3.", is, sff, at)
				}
				if l.nm[1][3] == nil {
					l.nm[1][3] = readLifeBarName("p4.", is, sff, at)
				}
				if l.nm[1][4] == nil {
					l.nm[1][4] = readLifeBarName("p5.", is, sff, at)
				}
				if l.nm[1][5] == nil {
					l.nm[1][5] = readLifeBarName("p6.", is, sff, at)
				}
				if l.nm[1][6] == nil {
					l.nm[1][6] = readLifeBarName("p7.", is, sff, at)
				}
				if l.nm[1][7] == nil {
					l.nm[1][7] = readLifeBarName("p8.", is, sff, at)
				}
			}
		case "turns ":
			subname = strings.ToLower(subname)
			switch {
			case len(subname) >= 7 && subname[:7] == "lifebar":
				if l.hb[2][0] == nil {
					l.hb[2][0] = readHealthBar("p1.", is, sff, at)
				}
				if l.hb[2][1] == nil {
					l.hb[2][1] = readHealthBar("p2.", is, sff, at)
				}
			case len(subname) >= 8 && subname[:8] == "powerbar":
				if l.pb[2][0] == nil {
					l.pb[2][0] = readPowerBar("p1.", is, sff, at, l.snd)
				}
				if l.pb[2][1] == nil {
					l.pb[2][1] = readPowerBar("p2.", is, sff, at, l.snd)
				}
			case len(subname) >= 4 && subname[:4] == "face":
				if l.fa[2][0] == nil {
					l.fa[2][0] = readLifeBarFace("p1.", is, sff, at)
				}
				if l.fa[2][1] == nil {
					l.fa[2][1] = readLifeBarFace("p2.", is, sff, at)
				}
			case len(subname) >= 4 && subname[:4] == "name":
				if l.nm[2][0] == nil {
					l.nm[2][0] = readLifeBarName("p1.", is, sff, at)
				}
				if l.nm[2][1] == nil {
					l.nm[2][1] = readLifeBarName("p2.", is, sff, at)
				}
			}
		case "tag ":
			subname = strings.ToLower(subname)
			switch {
			case len(subname) >= 7 && subname[:7] == "lifebar":
				if l.hb[3][0] == nil {
					l.hb[3][0] = readHealthBar("p1.", is, sff, at)
				}
				if l.hb[3][1] == nil {
					l.hb[3][1] = readHealthBar("p2.", is, sff, at)
				}
				if l.hb[3][2] == nil {
					l.hb[3][2] = readHealthBar("p3.", is, sff, at)
				}
				if l.hb[3][3] == nil {
					l.hb[3][3] = readHealthBar("p4.", is, sff, at)
				}
			case len(subname) >= 8 && subname[:8] == "powerbar":
				if l.pb[3][0] == nil {
					l.pb[3][0] = readPowerBar("p1.", is, sff, at, l.snd)
				}
				if l.pb[3][1] == nil {
					l.pb[3][1] = readPowerBar("p2.", is, sff, at, l.snd)
				}
			case len(subname) >= 4 && subname[:4] == "face":
				if l.fa[3][0] == nil {
					l.fa[3][0] = readLifeBarFace("p1.", is, sff, at)
				}
				if l.fa[3][1] == nil {
					l.fa[3][1] = readLifeBarFace("p2.", is, sff, at)
				}
				if l.fa[3][2] == nil {
					l.fa[3][2] = readLifeBarFace("p3.", is, sff, at)
				}
				if l.fa[3][3] == nil {
					l.fa[3][3] = readLifeBarFace("p4.", is, sff, at)
				}
				if l.fa[3][4] == nil {
					l.fa[3][4] = readLifeBarFace("p5.", is, sff, at)
				}
				if l.fa[3][5] == nil {
					l.fa[3][5] = readLifeBarFace("p6.", is, sff, at)
				}
				if l.fa[3][6] == nil {
					l.fa[3][6] = readLifeBarFace("p7.", is, sff, at)
				}
				if l.fa[3][7] == nil {
					l.fa[3][7] = readLifeBarFace("p8.", is, sff, at)
				}
			case len(subname) >= 4 && subname[:4] == "name":
				if l.nm[3][0] == nil {
					l.nm[3][0] = readLifeBarName("p1.", is, sff, at)
				}
				if l.nm[3][1] == nil {
					l.nm[3][1] = readLifeBarName("p2.", is, sff, at)
				}
				if l.nm[3][2] == nil {
					l.nm[3][2] = readLifeBarName("p3.", is, sff, at)
				}
				if l.nm[3][3] == nil {
					l.nm[3][3] = readLifeBarName("p4.", is, sff, at)
				}
				if l.nm[3][4] == nil {
					l.nm[3][4] = readLifeBarName("p5.", is, sff, at)
				}
				if l.nm[3][5] == nil {
					l.nm[3][5] = readLifeBarName("p6.", is, sff, at)
				}
				if l.nm[3][6] == nil {
					l.nm[3][6] = readLifeBarName("p7.", is, sff, at)
				}
				if l.nm[3][7] == nil {
					l.nm[3][7] = readLifeBarName("p8.", is, sff, at)
				}
			}
		case "simul_3p ", "simul_4p ", "tag_3p ", "tag_4p ":
			i := 4
			switch name {
			case "simul_4p ":
				i = 5
			case "tag_3p ":
				i = 6
			case "tag_4p ":
				i = 7
			}
			subname = strings.ToLower(subname)
			switch {
			case len(subname) >= 7 && subname[:7] == "lifebar":
				if l.hb[i][0] == nil {
					l.hb[i][0] = readHealthBar("p1.", is, sff, at)
				}
				if l.hb[i][1] == nil {
					l.hb[i][1] = readHealthBar("p2.", is, sff, at)
				}
				if l.hb[i][2] == nil {
					l.hb[i][2] = readHealthBar("p3.", is, sff, at)
				}
				if l.hb[i][3] == nil {
					l.hb[i][3] = readHealthBar("p4.", is, sff, at)
				}
				if l.hb[i][4] == nil {
					l.hb[i][4] = readHealthBar("p5.", is, sff, at)
				}
				if l.hb[i][5] == nil {
					l.hb[i][5] = readHealthBar("p6.", is, sff, at)
				}
				if i == 5 || i == 7 {
					if l.hb[i][6] == nil {
						l.hb[i][6] = readHealthBar("p7.", is, sff, at)
					}
					if l.hb[i][7] == nil {
						l.hb[i][7] = readHealthBar("p8.", is, sff, at)
					}
				}
			case len(subname) >= 8 && subname[:8] == "powerbar":
				if l.pb[i][0] == nil {
					l.pb[i][0] = readPowerBar("p1.", is, sff, at, l.snd)
				}
				if l.pb[i][1] == nil {
					l.pb[i][1] = readPowerBar("p2.", is, sff, at, l.snd)
				}
				if l.pb[i][2] == nil {
					l.pb[i][2] = readPowerBar("p3.", is, sff, at, l.snd)
				}
				if l.pb[i][3] == nil {
					l.pb[i][3] = readPowerBar("p4.", is, sff, at, l.snd)
				}
				if l.pb[i][4] == nil {
					l.pb[i][4] = readPowerBar("p5.", is, sff, at, l.snd)
				}
				if l.pb[i][5] == nil {
					l.pb[i][5] = readPowerBar("p6.", is, sff, at, l.snd)
				}
				if i == 5 || i == 7 {
					if l.pb[i][6] == nil {
						l.pb[i][6] = readPowerBar("p7.", is, sff, at, l.snd)
					}
					if l.pb[i][7] == nil {
						l.pb[i][7] = readPowerBar("p8.", is, sff, at, l.snd)
					}
				}
			}
		case "winicon":
			if l.wi[0] == nil {
				l.wi[0] = readLifeBarWinIcon("p1.", is, sff, at)
			}
			if l.wi[1] == nil {
				l.wi[1] = readLifeBarWinIcon("p2.", is, sff, at)
			}
		case "time":
			if l.ti == nil {
				l.ti = readLifeBarTime(is, sff, at)
			}
		case "combo":
			if l.co == nil {
				l.co = readLifeBarCombo(is)
			}
		case "round":
			if l.ro == nil {
				l.ro = readLifeBarRound(is, sff, at, l.snd, l.fnt[:])
			}
		case "challenger":
			if l.ch == nil {
				l.ch = readLifeBarChallenger(is, sff, at, l.snd, l.fnt[:])
			}
		}
	}
	for k, v := range missing {
		if strings.Contains(k, "lifebar") {
			for i := 3; i < len(l.hb); i++ {
				if i == v {
					for j, d := range l.hb[1] {
						l.hb[i][j] = d
					}
				}
			}
		} else if strings.Contains(k, "powerbar") {
			for i := 1; i < len(l.pb); i++ {
				if i == v {
					for j, d := range l.pb[0] {
						l.pb[i][j] = d
					}
				}
			}
		} else if strings.Contains(k, "tag face") {
			for j, d := range l.fa[1] {
				l.fa[3][j] = d
			}
		} else if strings.Contains(k, "tag name") {
			for j, d := range l.nm[1] {
				l.nm[3][j] = d
			}
		}
	}
	//BGDef
	i = 0
	defmap := make(map[string][]IniSection)
	for i < len(lines) {
		is, name, _ := ReadIniSection(lines, &i)
		if i := strings.IndexAny(name, " \t"); i >= 0 {
			if name[:i] == "bg" {
				defmap["bg"] = append(defmap["bg"], is)
			}
			//} else {
			//	defmap[name] = append(defmap[name], is)
		}
	}
	s := newBGDef(deffile)
	s.at = at
	s.sff = sff
	var bglink *backGround
	for _, bgsec := range defmap["bg"] {
		if len(s.bg) > 0 && s.bg[len(s.bg)-1].positionlink {
			bglink = s.bg[len(s.bg)-1]
		}
		s.bg = append(s.bg, readBackGround(bgsec, bglink,
			s.sff, s.at, 0, 0))
	}
	bgcdef := *newBgCtrl()
	i = 0
	for i < len(lines) {
		is, name, _ := ReadIniSection(lines, &i)
		if len(name) > 0 && name[len(name)-1] == ' ' {
			name = name[:len(name)-1]
		}
		switch name {
		case "bgctrldef":
			bgcdef.bg, bgcdef.looptime = nil, -1
			if ids := is.readI32CsvForStage("ctrlid"); len(ids) > 0 &&
				(len(ids) > 1 || ids[0] != -1) {
				kishutu := make(map[int32]bool)
				for _, id := range ids {
					if kishutu[id] {
						continue
					}
					bgcdef.bg = append(bgcdef.bg, s.getBg(id)...)
					kishutu[id] = true
				}
			} else {
				bgcdef.bg = append(bgcdef.bg, s.bg...)
			}
			is.ReadI32("looptime", &bgcdef.looptime)
			for key := range is {
				if strings.HasPrefix(key, "trigger") {
					tr, ok := is.readBGCTrigger(key)
					if ok {
						bgcdef.trigger = append(bgcdef.trigger, tr)
					}
				}
			}
		case "bgctrl":
			bgc := newBgCtrl()
			*bgc = bgcdef
			if ids := is.readI32CsvForStage("ctrlid"); len(ids) > 0 {
				bgc.bg = nil
				if len(ids) > 1 || ids[0] != -1 {
					kishutu := make(map[int32]bool)
					for _, id := range ids {
						if kishutu[id] {
							continue
						}
						bgc.bg = append(bgc.bg, s.getBg(id)...)
						kishutu[id] = true
					}
				} else {
					bgc.bg = append(bgc.bg, s.bg...)
				}
			}
			bgc.read(is, len(s.bgc))
			s.bgc = append(s.bgc, *bgc)
		}
	}
	//s.localscl = float32(sys.gameWidth) / float32(sys.cam.localcoord[0])
	l.bgdef = s
	return l, nil
}
func (l *Lifebar) step() {
	for ti := range sys.tmode {
		for i := ti; i < l.num[0][ti]; i += 2 {
			l.hb[l.ref[0][ti]][i].step(float32(sys.chars[i][0].life)/
				float32(sys.chars[i][0].lifeMax), (sys.chars[i][0].getcombo != 0 ||
				sys.chars[i][0].ss.moveType == MT_H) &&
				!sys.chars[i][0].scf(SCF_over))
		}
	}
	for ti := range sys.tmode {
		for i := ti; i < l.num[1][ti]; i += 2 {
			l.pb[l.ref[1][ti]][i].step(float32(sys.chars[i][0].power)/
				float32(sys.chars[i][0].powerMax), sys.chars[i][0].power/1000)
		}
	}
	for ti := range sys.tmode {
		for i := ti; i < l.num[2][ti]; i += 2 {
			l.fa[l.ref[2][ti]][i].step()
		}
	}
	for ti := range sys.tmode {
		for i := ti; i < l.num[3][ti]; i += 2 {
			l.nm[l.ref[3][ti]][i].step()
		}
	}
	for i := range l.wi {
		l.wi[i].step(sys.wins[i])
	}
	l.ti.step()
	cb := [2]int32{}
	for i, ch := range sys.chars {
		for _, c := range ch {
			cb[^i&1] = Min(999, Max(c.getcombo, cb[^i&1]))
		}
	}
	l.co.step(cb)
	l.ch.step()
}
func (l *Lifebar) reset() {
	for ti, tm := range sys.tmode {
		l.ref[0][ti] = int(tm)
		l.ref[1][ti] = int(tm)
		l.ref[2][ti] = int(tm)
		l.ref[3][ti] = int(tm)
		if tm == TM_Simul {
			if sys.tagMode[ti] && sys.numSimul[ti] == 2 { //Tag 2P
				l.ref[0][ti] = 3
				l.ref[1][ti] = 3
				l.ref[2][ti] = 3
				l.ref[3][ti] = 3
			} else if sys.tagMode[ti] { //Tag 3P/4P
				l.ref[0][ti] = int(sys.numSimul[ti]) + 3
				l.ref[1][ti] = 3
				l.ref[2][ti] = 3
				l.ref[3][ti] = 3
			} else if sys.numSimul[ti] > 2 { //Simul 3P/4P
				l.ref[0][ti] = int(sys.numSimul[ti]) + 1
				l.ref[1][ti] = int(sys.numSimul[ti]) + 1
			}
		}
		l.num[0][ti] = len(l.hb[l.ref[0][ti]])
		l.num[1][ti] = len(l.pb[l.ref[1][ti]])
		l.num[2][ti] = len(l.fa[l.ref[2][ti]])
		l.num[3][ti] = len(l.nm[l.ref[3][ti]])
		if tm == TM_Simul {
			l.num[0][ti] = int(sys.numSimul[ti]) * 2
			if sys.powerShare[ti] {
				l.num[1][ti] = 2
			} else if !sys.tagMode[ti] {
				l.num[1][ti] = int(sys.numSimul[ti]) * 2
			}
			l.num[2][ti] = int(sys.numSimul[ti]) * 2
			l.num[3][ti] = int(sys.numSimul[ti]) * 2
		}
	}
	for _, hb := range l.hb {
		for i := range hb {
			hb[i].reset()
		}
	}
	for _, pb := range l.pb {
		for i := range pb {
			pb[i].reset()
		}
	}
	for _, fa := range l.fa {
		for i := range fa {
			fa[i].reset()
		}
	}
	for _, nm := range l.nm {
		for i := range nm {
			nm[i].reset()
		}
	}
	for i := range l.wi {
		l.wi[i].reset()
	}
	l.ti.reset()
	l.co.reset()
	l.ro.reset()
	l.ch.reset()
	l.bgdef.reset()
}
func (l *Lifebar) draw(layerno int16) {
	if !sys.statusDraw {
		return
	}
	if !sys.sf(GSF_nobardisplay) && sys.gameMode != "demo" {
		for ti := range sys.tmode {
			for i := ti; i < l.num[0][ti]; i += 2 {
				l.hb[l.ref[0][ti]][i].bgDraw(layerno)
			}
		}
		for ti := range sys.tmode {
			for i := ti; i < l.num[0][ti]; i += 2 {
				l.hb[l.ref[0][ti]][i].draw(layerno, float32(sys.chars[i][0].life)/
					float32(sys.chars[i][0].lifeMax))
			}
		}
		for ti := range sys.tmode {
			for i := ti; i < l.num[1][ti]; i += 2 {
				l.pb[l.ref[1][ti]][i].bgDraw(layerno)
			}
		}
		for ti := range sys.tmode {
			for i := ti; i < l.num[1][ti]; i += 2 {
				l.pb[l.ref[1][ti]][i].draw(layerno, float32(sys.chars[i][0].power)/
					float32(sys.chars[i][0].powerMax), sys.chars[i][0].power/1000,
					l.fnt[:], float32(sys.chars[i][0].power), float32(sys.chars[i][0].powerMax))
			}
		}
		for ti := range sys.tmode {
			for i := ti; i < l.num[2][ti]; i += 2 {
				l.fa[l.ref[2][ti]][i].bgDraw(layerno)
			}
		}
		for ti := range sys.tmode {
			for i := ti; i < l.num[2][ti]; i += 2 {
				if fspr := l.fa[l.ref[2][ti]][i].face; fspr != nil {
					pfx := sys.chars[i][0].getPalfx()
					sys.cgi[i].sff.palList.SwapPalMap(&pfx.remap)
					fspr.Pal = nil
					fspr.Pal = fspr.GetPal(&sys.cgi[i].sff.palList)
					sys.cgi[i].sff.palList.SwapPalMap(&pfx.remap)
					l.fa[l.ref[2][ti]][i].draw(layerno, pfx, i == sys.superplayer)
				}
			}
		}
		for ti := range sys.tmode {
			for i := ti; i < l.num[3][ti]; i += 2 {
				l.nm[l.ref[3][ti]][i].bgDraw(layerno)
			}
		}
		for ti := range sys.tmode {
			for i := ti; i < l.num[3][ti]; i += 2 {
				l.nm[l.ref[3][ti]][i].draw(layerno, l.fnt[:], sys.cgi[i].displayname)
			}
		}
		l.ti.bgDraw(layerno)
		l.ti.draw(layerno, l.fnt[:])
		for i := range l.wi {
			l.wi[i].draw(layerno, l.fnt[:])
		}
		if layerno == 0 {
			l.bgdef.draw(false, 0, 0, 1)
		} else if layerno == 1 {
			l.bgdef.draw(true, 0, 0, 1)
		}
	}
	l.co.draw(layerno, l.fnt[:])
	if sys.challenger > 0 && l.ch.challenger.displaytime > l.ch.cnt {
		l.ch.bgDraw(layerno)
		l.ch.draw(layerno, l.fnt[:])
	}
}

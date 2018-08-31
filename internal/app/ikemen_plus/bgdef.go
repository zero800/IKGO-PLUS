package main

import (
	"strings"
	"unicode"
)

func (bgct *bgcTimeLine) stepBGDef(s *BGDef) {
	if len(bgct.line) > 0 && bgct.line[0].waitTime <= 0 {
		for _, b := range bgct.line[0].bgc {
			for i, a := range bgct.al {
				if b.idx < a.idx {
					bgct.al = append(bgct.al, nil)
					copy(bgct.al[i+1:], bgct.al[i:])
					bgct.al[i] = b
					b = nil
					break
				}
			}
			if b != nil {
				bgct.al = append(bgct.al, b)
			}
		}
		bgct.line = bgct.line[1:]
	}
	if len(bgct.line) > 0 {
		bgct.line[0].waitTime--
	}
	var el []*bgCtrl
	for i := 0; i < len(bgct.al); {
		s.runBgCtrl(bgct.al[i])
		if bgct.al[i].currenttime > bgct.al[i].endtime {
			el = append(el, bgct.al[i])
			bgct.al = append(bgct.al[:i], bgct.al[i+1:]...)
			continue
		}
		i++
	}
	for _, b := range el {
		bgct.add(b)
	}
}

type BGDef struct {
	def        string
	localcoord [2]float32
	sffloc     string
	sff        *Sff
	at         AnimationTable
	bg         []*backGround
	bgc        []bgCtrl
	bgct       bgcTimeLine
	bga        bgAction
	resetbg    bool
	localscl   float32
	scale      [2]float32
}

func newBGDef(def string) *BGDef {
	s := &BGDef{def: def, localcoord: [...]float32{320, 240}, resetbg: true, localscl: 1, scale: [...]float32{1, 1}}
	return s
}

func loadBGDef(def string, bgname string, sffloc string) (int, error) {
	s := newBGDef(def)
	str, err := LoadText(def)
	if err != nil {
		return -1, err
	}
	s.sff = &Sff{}
	lines, i := SplitAndTrim(str, "\n"), 0
	defmap := make(map[string][]IniSection)
	for i < len(lines) {
		is, name, _ := ReadIniSection(lines, &i)
		if i := strings.IndexAny(name, " \t"); i >= 0 {
			if name[:i] == bgname {
				defmap[bgname] = append(defmap[bgname], is)
			}
		} else {
			defmap[name] = append(defmap[name], is)
		}
	}
	i = 0
	if sec := defmap["info"]; len(sec) > 0 {
		sec[0].readF32ForStage("localcoord", &s.localcoord[0], &s.localcoord[1])
	}
	var ok, skipat bool
	var filename string
	bgnum := -1
	if sffloc != "" {
		filename = sffloc
	} else if sec := defmap["files"]; len(sec) > 0 {
		if sec[0].LoadFile("spr", def, func(filename string) error {
			filename = strings.Replace(filename, "\\", "/", -1)
			return nil
		}); err != nil {
			return -1, err
		}
	}
	for j := 0; j < len(sys.bgdef); j++ {
		if !ok && sys.bgdef[j].sffloc == filename {
			*s.sff = *sys.bgdef[j].sff
			bgnum = j
			ok = true
		}
		if sys.bgdef[j].def == def && ok {
			skipat = true
			break
		}
	}
	if !ok { //skip loadSFF if already loaded
		sff, err := loadSff(filename, false)
		if err != nil {
			return -1, err
		}
		*s.sff = *sff
	}
	s.sffloc = filename
	if skipat { //skip ReadAnimationTable if already parsed
		s.at = sys.bgdef[bgnum].at
	} else {
		s.at = ReadAnimationTable(s.sff, lines, &i)
	}
	var bglink *backGround
	for _, bgsec := range defmap[bgname] {
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
		case bgname + "ctrldef":
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
		case bgname + "ctrl":
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
	sys.bgdef = append(sys.bgdef, s)
	return len(sys.bgdef) - 1, nil
}
func (s *BGDef) getBg(id int32) (bg []*backGround) {
	if id >= 0 {
		for _, b := range s.bg {
			if b.id == id {
				bg = append(bg, b)
			}
		}
	}
	return
}
func (s *BGDef) runBgCtrl(bgc *bgCtrl) {
	bgc.currenttime++
	switch bgc._type {
	case BT_Anim:
		a := s.at.get(bgc.v[0])
		if a != nil {
			for i := range bgc.bg {
				bgc.bg[i].actionno = bgc.v[0]
				bgc.bg[i].anim = *a
			}
		}
	case BT_Visible:
		for i := range bgc.bg {
			bgc.bg[i].visible = bgc.v[0] != 0
		}
	case BT_Enable:
		for i := range bgc.bg {
			bgc.bg[i].visible, bgc.bg[i].active = bgc.v[0] != 0, bgc.v[0] != 0
		}
	case BT_PosSet:
		for i := range bgc.bg {
			if bgc.xEnable() {
				bgc.bg[i].bga.pos[0] = bgc.x
			}
			if bgc.yEnable() {
				bgc.bg[i].bga.pos[1] = bgc.y
			}
		}
		if bgc.positionlink {
			if bgc.xEnable() {
				s.bga.pos[0] = bgc.x
			}
			if bgc.yEnable() {
				s.bga.pos[1] = bgc.y
			}
		}
	case BT_PosAdd:
		for i := range bgc.bg {
			if bgc.xEnable() {
				bgc.bg[i].bga.pos[0] += bgc.x
			}
			if bgc.yEnable() {
				bgc.bg[i].bga.pos[1] += bgc.y
			}
		}
		if bgc.positionlink {
			if bgc.xEnable() {
				s.bga.pos[0] += bgc.x
			}
			if bgc.yEnable() {
				s.bga.pos[1] += bgc.y
			}
		}
	case BT_SinX, BT_SinY:
		ii := Btoi(bgc._type == BT_SinY)
		if bgc.v[0] == 0 {
			bgc.v[1] = 0
		}
		a := float32(bgc.v[2]) / 360
		st := int32((a - float32(int32(a))) * float32(bgc.v[1]))
		if st < 0 {
			st += Abs(bgc.v[1])
		}
		for i := range bgc.bg {
			bgc.bg[i].bga.radius[ii] = bgc.x
			bgc.bg[i].bga.sinlooptime[ii] = bgc.v[1]
			bgc.bg[i].bga.sintime[ii] = st
		}
		if bgc.positionlink {
			s.bga.radius[ii] = bgc.x
			s.bga.sinlooptime[ii] = bgc.v[1]
			s.bga.sintime[ii] = st
		}
	case BT_VelSet:
		for i := range bgc.bg {
			if bgc.xEnable() {
				bgc.bg[i].bga.vel[0] = bgc.x
			}
			if bgc.yEnable() {
				bgc.bg[i].bga.vel[1] = bgc.y
			}
		}
		if bgc.positionlink {
			if bgc.xEnable() {
				s.bga.vel[0] = bgc.x
			}
			if bgc.yEnable() {
				s.bga.vel[1] = bgc.y
			}
		}
	case BT_VelAdd:
		for i := range bgc.bg {
			if bgc.xEnable() {
				bgc.bg[i].bga.vel[0] += bgc.x
			}
			if bgc.yEnable() {
				bgc.bg[i].bga.vel[1] += bgc.y
			}
		}
		if bgc.positionlink {
			if bgc.xEnable() {
				s.bga.vel[0] += bgc.x
			}
			if bgc.yEnable() {
				s.bga.vel[1] += bgc.y
			}
		}
	}
}
func (s *BGDef) action() {
	s.bgct.stepBGDef(s)
	s.bga.action()
	link := 0
	for i, b := range s.bg {
		s.bg[i].bga.action()
		if i > 0 && b.positionlink {
			s.bg[i].bga.offset[0] += s.bg[link].bga.sinoffset[0]
			s.bg[i].bga.offset[1] += s.bg[link].bga.sinoffset[1]
		} else {
			link = i
		}
		if b.active {
			s.bg[i].anim.Action()
		}
	}
}
func (s *BGDef) draw(top bool, x, y, scl float32) {
	if !top {
		s.action()
	}
	x, y = x/s.localscl, y/s.localscl
	bgscl := float32(1)
	pos := [...]float32{x, y}
	for _, b := range s.bg {
		if b.visible && b.toplayer == top && b.anim.spr != nil {
			b.draw(pos, scl, bgscl, s.localscl, s.scale, 0)
		}
	}
}
func (s *BGDef) reset() {
	s.bga.clear()
	for i := range s.bg {
		s.bg[i].reset()
	}
	for i := range s.bgc {
		s.bgc[i].currenttime = 0
	}
	s.bgct.clear()
	for i := len(s.bgc) - 1; i >= 0; i-- {
		s.bgct.add(&s.bgc[i])
	}
}

type BgcTrigger int32

const (
	BR_Null BgcTrigger = iota
	BR_AILevel
	BR_Alive
	BR_Anim
	BR_AnimExist
	BR_AuthorName
	BR_BackEdgeBodyDist
	BR_BackEdgeDist
	BR_CameraPosX
	BR_CameraPosY
	BR_Ctrl
	BR_DisplayName
	BR_DrawGame
	BR_Facing
	BR_FinishType
	BR_FrontEdgeBodyDist
	BR_FrontEdgeDist
	BR_FVar
	BR_GameMode
	BR_GameTime
	BR_HitCount
	BR_HitFall
	BR_HitOver
	BR_HitPauseTime
	BR_HitShakeOver
	BR_HitVelX
	BR_HitVelY
	BR_InGuardDist
	BR_Life
	BR_LifeMax
	BR_LifePercentage
	BR_Lose
	BR_LoseKO
	BR_LoseTime
	BR_MatchNo
	BR_MatchOver
	BR_MoveContact
	BR_MoveGuarded
	BR_MoveHit
	BR_MoveReversed
	BR_MoveType
	BR_Name
	BR_NumEnemy
	BR_NumPartner
	BR_PalNo
	BR_PosX
	BR_PosY
	BR_Power
	BR_PowerLevel
	BR_PowerMax
	BR_PowerPercentage
	BR_PrevStateNo
	BR_RoundNo
	BR_RoundType
	BR_RoundsExisted
	BR_RoundState
	BR_ScreenPosX
	BR_ScreenPosY
	BR_SelfAnimExist
	BR_StateNo
	BR_StateType
	BR_StageName
	BR_StageDisplayName
	BR_StageAuthor
	BR_SysFVar
	BR_SysVar
	BR_TagMode
	BR_TeamMode
	BR_Time
	BR_UniqHitCount
	BR_Var
	BR_VelX
	BR_VelY
	BR_Win
	BR_WinKO
	BR_WinTeam
	BR_WinTime
	BR_WinType
	BR_WinPerfect
)

type BgcTriggerOp int32

const (
	OP_Null BgcTriggerOp = iota
	OP_Equal
	OP_NotEqual
	OP_Greater
	OP_GreaterOrEqual
	OP_Less
	OP_LessOrEqual
)

type bgcTrigger struct {
	key BgcTrigger
	op  BgcTriggerOp
	i32 int32
	b   bool
	s   string
	px  int
	vx  int32
}

func (is IniSection) readBGCTrigger(name string) (*bgcTrigger, bool) {
	tr := &bgcTrigger{}
	if str := strings.ToLower(is[name]); len(str) > 0 {
		val := 0
		for i, s := range strings.Split(str, ",") {
			if s = strings.TrimLeftFunc(s, unicode.IsSpace); len(s) > 0 {
				switch i {
				case 0: //name
					switch s {
					case "ailevel":
						tr.key = BR_AILevel
					case "alive":
						tr.key = BR_Alive
						val = 1
					case "anim":
						tr.key = BR_Anim
					case "animexist":
						tr.key = BR_AnimExist
					case "authorname":
						tr.key = BR_AuthorName
						val = 2
					case "backedgebodydist":
						tr.key = BR_BackEdgeBodyDist
					case "backedgedist":
						tr.key = BR_BackEdgeDist
					case "cameraposx":
						tr.key = BR_CameraPosX
					case "cameraposy":
						tr.key = BR_CameraPosY
					case "ctrl":
						tr.key = BR_Ctrl
						val = 1
					case "displayname":
						tr.key = BR_DisplayName
						val = 2
					case "drawgame":
						tr.key = BR_DrawGame
						val = 1
					case "facing":
						tr.key = BR_Facing
					case "finishtype":
						tr.key = BR_FinishType
						val = 2
					case "frontedgebodydist":
						tr.key = BR_FrontEdgeBodyDist
					case "frontedgedist":
						tr.key = BR_FrontEdgeDist
					case "fvar":
						tr.key = BR_FVar
					case "gamemode":
						tr.key = BR_GameMode
						val = 2
					case "gametime":
						tr.key = BR_GameTime
					case "hitcount":
						tr.key = BR_HitCount
					case "hitfall":
						tr.key = BR_HitFall
						val = 1
					case "hitover":
						tr.key = BR_HitOver
						val = 1
					case "hitpausetime":
						tr.key = BR_HitPauseTime
					case "hitshakeover":
						tr.key = BR_HitShakeOver
						val = 1
					case "hitvelx":
						tr.key = BR_HitVelX
					case "hitvely":
						tr.key = BR_HitVelY
					case "inguarddist":
						tr.key = BR_InGuardDist
						val = 1
					case "life":
						tr.key = BR_Life
					case "lifemax":
						tr.key = BR_LifeMax
					case "lifepercentage":
						tr.key = BR_LifePercentage
					case "lose":
						tr.key = BR_Lose
						val = 1
					case "loseko":
						tr.key = BR_LoseKO
						val = 1
					case "losetime":
						tr.key = BR_LoseTime
						val = 1
					case "matchno":
						tr.key = BR_MatchNo
					case "matchover":
						tr.key = BR_MatchOver
						val = 1
					case "movecontact":
						tr.key = BR_MoveContact
					case "moveguarded":
						tr.key = BR_MoveGuarded
					case "movehit":
						tr.key = BR_MoveHit
					case "movereversed":
						tr.key = BR_MoveReversed
					case "movetype":
						tr.key = BR_MoveType
						val = 2
					case "name":
						tr.key = BR_Name
						val = 2
					case "numenemy":
						tr.key = BR_NumEnemy
					case "numpartner":
						tr.key = BR_NumPartner
					case "palno":
						tr.key = BR_PalNo
					case "posx":
						tr.key = BR_PosX
					case "posy":
						tr.key = BR_PosY
					case "power":
						tr.key = BR_Power
					case "powerlevel":
						tr.key = BR_PowerLevel
					case "powermax":
						tr.key = BR_PowerMax
					case "powerpercentage":
						tr.key = BR_PowerPercentage
					case "prevstateno":
						tr.key = BR_PrevStateNo
					case "roundno":
						tr.key = BR_RoundNo
					case "roundtype":
						tr.key = BR_RoundType
						val = 2
					case "roundsexisted":
						tr.key = BR_RoundsExisted
					case "roundstate":
						tr.key = BR_RoundState
					case "screenposx":
						tr.key = BR_ScreenPosX
					case "screenposy":
						tr.key = BR_ScreenPosY
					case "selfanimexist":
						tr.key = BR_SelfAnimExist
					case "stateno":
						tr.key = BR_StateNo
					case "statetype":
						tr.key = BR_StateType
						val = 2
					case "stagename":
						tr.key = BR_StageName
						val = 2
					case "stagedisplayname":
						tr.key = BR_StageDisplayName
						val = 2
					case "stageauthor":
						tr.key = BR_StageAuthor
						val = 2
					case "sysfvar":
						tr.key = BR_SysFVar
					case "sysvar":
						tr.key = BR_SysVar
					case "tagmode":
						tr.key = BR_TagMode
						val = 1
					case "teammode":
						tr.key = BR_TeamMode
					case "time":
						tr.key = BR_Time
					case "uniqhitcount":
						tr.key = BR_UniqHitCount
					case "var":
						tr.key = BR_Var
					case "velx":
						tr.key = BR_VelX
					case "vely":
						tr.key = BR_VelY
					case "win":
						tr.key = BR_Win
						val = 1
					case "winko":
						tr.key = BR_WinKO
						val = 1
					case "winteam":
						tr.key = BR_WinTeam
					case "wintime":
						tr.key = BR_WinTime
						val = 1
					case "wintype":
						tr.key = BR_WinType
						val = 2
					case "winperfect":
						tr.key = BR_WinPerfect
						val = 1
					default:
						return nil, false
					}
				case 1: //operator
					switch s {
					case "=":
						tr.op = OP_Equal
					case "==":
						tr.op = OP_Equal
					case "!=":
						tr.op = OP_NotEqual
					case ">":
						tr.op = OP_Greater
					case ">=":
						tr.op = OP_GreaterOrEqual
					case "<":
						tr.op = OP_Less
					case "<=":
						tr.op = OP_LessOrEqual
					default:
						return nil, false
					}
				case 2: //value
					switch val {
					case 1: //bool
						switch s {
						case "true":
							tr.b = true
						default:
							tr.b = false
						}
					case 2: //string
						tr.s = s
					default: //int32
						tr.i32 = Atoi(s)
					}
				case 3: //player num
					tr.px = int(Atoi(s))
				case 4: //variable num
					tr.vx = Atoi(s)
				default:
					return nil, false
				}
			}
			if strings.IndexFunc(s, unicode.IsSpace) >= 0 {
				break
			}
		}
	}
	return tr, true
}

func bgcTriggerBool(t *bgcTrigger) bool {
	val := 0
	var ln, rn int32
	var lb, rb bool
	var ls, rs string
	switch t.key {
	case BR_AILevel:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = sys.chars[t.px-1][0].aiLevel(), t.i32
		} else {
			return false
		}
	case BR_Alive:
		if len(sys.chars[t.px-1]) > 0 {
			lb, rb = sys.chars[t.px-1][0].alive(), t.b
			val = 1
		} else {
			return false
		}
	case BR_Anim:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = sys.chars[t.px-1][0].animNo, t.i32
		} else {
			return false
		}
	case BR_AnimExist:
		if len(sys.chars[t.px-1]) > 0 {
			lb, rb = sys.chars[t.px-1][0].animExist(sys.chars[t.px-1][0], BytecodeInt(t.i32)).ToB(), true
			val = 1
		} else {
			return false
		}
	case BR_AuthorName:
		if len(sys.chars[t.px-1]) > 0 {
			ls, rs = strings.ToLower(sys.chars[t.px-1][0].gi().author), t.s
			val = 2
		} else {
			return false
		}
	case BR_BackEdgeBodyDist:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = int32(sys.chars[t.px-1][0].backEdgeBodyDist()), t.i32
		} else {
			return false
		}
	case BR_BackEdgeDist:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = int32(sys.chars[t.px-1][0].backEdgeDist()), t.i32
		} else {
			return false
		}
	case BR_CameraPosX:
		ln, rn = int32(sys.cam.Pos[0]), t.i32
	case BR_CameraPosY:
		ln, rn = int32(sys.cam.Pos[1]), t.i32
	case BR_Ctrl:
		if len(sys.chars[t.px-1]) > 0 {
			lb, rb = sys.chars[t.px-1][0].ctrl(), t.b
			val = 1
		} else {
			return false
		}
	case BR_DisplayName:
		if len(sys.chars[t.px-1]) > 0 {
			ls, rs = strings.ToLower(sys.chars[t.px-1][0].gi().displayname), t.s
			val = 2
		} else {
			return false
		}
	case BR_DrawGame:
		if len(sys.chars[t.px-1]) > 0 {
			lb, rb = sys.chars[t.px-1][0].drawgame(), t.b
			val = 1
		} else {
			return false
		}
	case BR_Facing:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = int32(sys.chars[t.px-1][0].facing), t.i32
		} else {
			return false
		}
	case BR_FinishType:
		switch sys.finish {
		case FT_KO: //KO
			ls = "ko"
		case FT_DKO: //Double KO
			ls = "dko"
		case FT_TO: //Time Over
			ls = "to"
		case FT_TODraw: //Time Over Draw
			ls = "todraw"
		default:
			ls = ""
		}
		rs = t.s
		val = 2
	case BR_FrontEdgeBodyDist:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = int32(sys.chars[t.px-1][0].frontEdgeBodyDist()), t.i32
		} else {
			return false
		}
	case BR_FrontEdgeDist:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = int32(sys.chars[t.px-1][0].frontEdgeDist()), t.i32
		} else {
			return false
		}
	case BR_FVar:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = sys.chars[t.px-1][0].fvarGet(t.vx).ToI(), t.i32
		} else {
			return false
		}
	case BR_GameMode:
		ls, rs = strings.ToLower(sys.gameMode), t.s
		val = 2
	case BR_GameTime:
		ln, rn = sys.gameTime, t.i32
	case BR_HitCount:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = sys.chars[t.px-1][0].hitCount, t.i32
		} else {
			return false
		}
	case BR_HitFall:
		if len(sys.chars[t.px-1]) > 0 {
			lb, rb = sys.chars[t.px-1][0].ghv.fallf, t.b
			val = 1
		} else {
			return false
		}
	case BR_HitOver:
		if len(sys.chars[t.px-1]) > 0 {
			lb, rb = sys.chars[t.px-1][0].hitOver(), t.b
			val = 1
		} else {
			return false
		}
	case BR_HitPauseTime:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = sys.chars[t.px-1][0].hitPauseTime, t.i32
		} else {
			return false
		}
	case BR_HitShakeOver:
		if len(sys.chars[t.px-1]) > 0 {
			lb, rb = sys.chars[t.px-1][0].hitShakeOver(), t.b
			val = 1
		} else {
			return false
		}
	case BR_HitVelX:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = int32(sys.chars[t.px-1][0].hitVelX()), t.i32
		} else {
			return false
		}
	case BR_HitVelY:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = int32(sys.chars[t.px-1][0].hitVelY()), t.i32
		} else {
			return false
		}
	case BR_InGuardDist:
		if len(sys.chars[t.px-1]) > 0 {
			lb, rb = sys.chars[t.px-1][0].inguarddist, t.b
			val = 1
		} else {
			return false
		}
	case BR_Life:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = sys.chars[t.px-1][0].life, t.i32
		} else {
			return false
		}
	case BR_LifeMax:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = sys.chars[t.px-1][0].lifeMax, t.i32
		} else {
			return false
		}
	case BR_LifePercentage:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = int32(float32(sys.chars[t.px-1][0].life)/float32(sys.chars[t.px-1][0].lifeMax)*100), t.i32
		} else {
			return false
		}
	case BR_Lose:
		if len(sys.chars[t.px-1]) > 0 {
			lb, rb = sys.chars[t.px-1][0].lose(), t.b
			val = 1
		} else {
			return false
		}
	case BR_LoseKO:
		if len(sys.chars[t.px-1]) > 0 {
			lb, rb = sys.chars[t.px-1][0].loseKO(), t.b
			val = 1
		} else {
			return false
		}
	case BR_LoseTime:
		if len(sys.chars[t.px-1]) > 0 {
			lb, rb = sys.chars[t.px-1][0].loseTime(), t.b
			val = 1
		} else {
			return false
		}
	case BR_MatchNo:
		ln, rn = sys.match, t.i32
	case BR_MatchOver:
		lb, rb = sys.matchOver(), t.b
		val = 1
	case BR_MoveContact:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = sys.chars[t.px-1][0].moveContact(), t.i32
		} else {
			return false
		}
	case BR_MoveGuarded:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = sys.chars[t.px-1][0].moveGuarded(), t.i32
		} else {
			return false
		}
	case BR_MoveHit:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = sys.chars[t.px-1][0].moveHit(), t.i32
		} else {
			return false
		}
	case BR_MoveReversed:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = sys.chars[t.px-1][0].moveReversed(), t.i32
		} else {
			return false
		}
	case BR_MoveType:
		if len(sys.chars[t.px-1]) > 0 {
			switch sys.chars[t.px-1][0].ss.moveType {
			case MT_I: //Idle
				ls = "i"
			case MT_A: //Attack
				ls = "a"
			case MT_H: //GetHit
				ls = "h"
			default:
				ls = ""
			}
			rs = t.s
			val = 2
		} else {
			return false
		}
	case BR_Name:
		if len(sys.chars[t.px-1]) > 0 {
			ls, rs = strings.ToLower(sys.chars[t.px-1][0].name), t.s
			val = 2
		} else {
			return false
		}
	case BR_NumEnemy:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = sys.chars[t.px-1][0].numEnemy(), t.i32
		} else {
			return false
		}
	case BR_NumPartner:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = sys.chars[t.px-1][0].numPartner(), t.i32
		} else {
			return false
		}
	case BR_PalNo:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = sys.chars[t.px-1][0].gi().palno, t.i32
		} else {
			return false
		}
	case BR_PosX:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = int32(sys.chars[t.px-1][0].pos[0]-sys.cam.Pos[0]), t.i32
		} else {
			return false
		}
	case BR_PosY:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = int32(sys.chars[t.px-1][0].pos[1]), t.i32
		} else {
			return false
		}
	case BR_Power:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = sys.chars[t.px-1][0].power, t.i32
		} else {
			return false
		}
	case BR_PowerLevel:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = sys.chars[t.px-1][0].power/1000, t.i32
		} else {
			return false
		}
	case BR_PowerMax:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = sys.chars[t.px-1][0].powerMax, t.i32
		} else {
			return false
		}
	case BR_PowerPercentage:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = int32(float32(sys.chars[t.px-1][0].power)/float32(sys.chars[t.px-1][0].powerMax)*100), t.i32
		} else {
			return false
		}
	case BR_PrevStateNo:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = sys.chars[t.px-1][0].ss.prevno, t.i32
		} else {
			return false
		}
	case BR_RoundNo:
		ln, rn = sys.round, t.i32
	case BR_RoundType:
		switch sys.roundType[t.px-1] {
		case RT_Normal: //Normal round (never finishes the match)
			ls = "normal"
		case RT_Deciding: //Deciding round (can end the match if pX looses)
			ls = "deciding"
		case RT_Final: //Final round (always ends the match)
			ls = "final"
		default:
			ls = ""
		}
		rs = t.s
		val = 2
	case BR_RoundsExisted:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = sys.chars[t.px-1][0].roundsExisted(), t.i32
		} else {
			return false
		}
	case BR_RoundState:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = sys.chars[t.px-1][0].roundState(), t.i32
		} else {
			return false
		}
	case BR_ScreenPosX:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = int32(sys.chars[t.px-1][0].screenPosX()), t.i32
		} else {
			return false
		}
	case BR_ScreenPosY:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = int32(sys.chars[t.px-1][0].screenPosY()), t.i32
		} else {
			return false
		}
	case BR_SelfAnimExist:
		if len(sys.chars[t.px-1]) > 0 {
			lb, rb = sys.chars[t.px-1][0].selfAnimExist(BytecodeInt(t.i32)).ToB(), true
			val = 1
		} else {
			return false
		}
	case BR_StateNo:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = sys.chars[t.px-1][0].ss.no, t.i32
		} else {
			return false
		}
	case BR_StateType:
		if len(sys.chars[t.px-1]) > 0 {
			switch sys.chars[t.px-1][0].ss.stateType {
			case ST_S: //Stand
				ls = "s"
			case ST_C: //Crouch
				ls = "c"
			case ST_A: //Air state-type
				ls = "a"
			case ST_L: //?
				ls = "l"
			default:
				ls = ""
			}
			rs = t.s
			val = 2
		} else {
			return false
		}
	case BR_StageName:
		ls, rs = strings.ToLower(sys.stage.name), t.s
		val = 2
	case BR_StageDisplayName:
		ls, rs = strings.ToLower(sys.stage.displayname), t.s
		val = 2
	case BR_StageAuthor:
		ls, rs = strings.ToLower(sys.stage.author), t.s
		val = 2
	case BR_SysFVar:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = sys.chars[t.px-1][0].sysFvarGet(t.vx).ToI(), t.i32
		} else {
			return false
		}
	case BR_SysVar:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = sys.chars[t.px-1][0].sysVarGet(t.vx).ToI(), t.i32
		} else {
			return false
		}
	case BR_TagMode:
		lb, rb = sys.tagMode[t.px-1], t.b
		val = 1
	case BR_TeamMode:
		ln, rn = int32(sys.tmode[t.px-1]), t.i32
	case BR_Time:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = sys.chars[t.px-1][0].ss.time, t.i32
		} else {
			return false
		}
	case BR_UniqHitCount:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = sys.chars[t.px-1][0].uniqHitCount, t.i32
		} else {
			return false
		}
	case BR_Var:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = sys.chars[t.px-1][0].varGet(t.vx).ToI(), t.i32
		} else {
			return false
		}
	case BR_VelX:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = int32(sys.chars[t.px-1][0].vel[0]), t.i32
		} else {
			return false
		}
	case BR_VelY:
		if len(sys.chars[t.px-1]) > 0 {
			ln, rn = int32(sys.chars[t.px-1][0].vel[1]), t.i32
		} else {
			return false
		}
	case BR_Win:
		if len(sys.chars[t.px-1]) > 0 {
			lb, rb = sys.chars[t.px-1][0].win(), t.b
			val = 1
		} else {
			return false
		}
	case BR_WinKO:
		if len(sys.chars[t.px-1]) > 0 {
			lb, rb = sys.chars[t.px-1][0].winKO(), t.b
			val = 1
		} else {
			return false
		}
	case BR_WinTeam:
		ln, rn = int32(sys.winTeam), t.i32
	case BR_WinTime:
		if len(sys.chars[t.px-1]) > 0 {
			lb, rb = sys.chars[t.px-1][0].winTime(), t.b
			val = 1
		} else {
			return false
		}
	case BR_WinType:
		switch sys.winType[t.px-1] {
		case WT_N: //Win by normal
			ls = "n"
		case WT_S: //Win by special
			ls = "s"
		case WT_H: //Win by hyper (super)
			ls = "h"
		case WT_C: //Win by cheese
			ls = "c"
		case WT_T: //Win by time over
			ls = "t"
		case WT_Throw: //Win by normal throw
			ls = "throw"
		case WT_Suicide: //Win by suicide
			ls = "suicide"
		case WT_Teammate: //Opponent beaten by his own teammate
			ls = "teammate"
		case WT_Perfect: //Win by perfect
			ls = "perfect"
		case WT_NumTypes:
		case WT_PN: //Win by normal (perfect)
			ls = "pn"
		case WT_PS: //Win by special (perfect)
			ls = "ps"
		case WT_PH: //Win by hyper (super) (perfect)
			ls = "ph"
		case WT_PC: //Win by cheese (perfect)
			ls = "pc"
		case WT_PT: //Win by time over (perfect)
			ls = "pt"
		case WT_PThrow: //Win by normal throw (perfect)
			ls = "pthrow"
		case WT_PSuicide: //Win by suicide (perfect)
			ls = "psuicide"
		case WT_PTeammate: //Opponent beaten by his own teammate (perfect)
			ls = "pteammate"
		default:
			ls = ""
		}
		rs = t.s
		val = 2
	case BR_WinPerfect:
		if len(sys.chars[t.px-1]) > 0 {
			lb, rb = sys.chars[t.px-1][0].winPerfect(), t.b
			val = 1
		} else {
			return false
		}
	}
	switch val {
	case 1: //bool
		switch t.op {
		case OP_Equal:
			if lb == rb {
				return true
			}
		case OP_NotEqual:
			if lb != rb {
				return true
			}
		}
	case 2: //string
		switch t.op {
		case OP_Equal:
			if ls == rs {
				return true
			}
		case OP_NotEqual:
			if ls != rs {
				return true
			}
		}
	default: //int32
		switch t.op {
		case OP_Equal:
			if ln == rn {
				return true
			}
		case OP_NotEqual:
			if ln != rn {
				return true
			}
		case OP_Greater:
			if ln > rn {
				return true
			}
		case OP_GreaterOrEqual:
			if ln >= rn {
				return true
			}
		case OP_Less:
			if ln < rn {
				return true
			}
		case OP_LessOrEqual:
			if ln <= rn {
				return true
			}
		}
	}
	return false
}

package main

import "testing"

func TestValidAnimFrameLine(t *testing.T) {

	lineTests := []struct {
		name string
		afl  AnimFrameLine
		want bool
	}{
		{name: "Empty", afl: AnimFrameLine{Line: ""}, want: false},
		{name: "Missing time", afl: AnimFrameLine{Line: "0,0, 0,0"}, want: false},
		{name: "Comments", afl: AnimFrameLine{Line: "0,0, 0,0, 5 ;comment"}, want: true},
		{name: "Pos & Transparency", afl: AnimFrameLine{Line: "0,0, 0,0, 10, , S"}, want: true},
		{name: "Addalpha", afl: AnimFrameLine{Line: "0,0, 0,0, 8, , AS249D200"}, want: true},
		{name: "Xscale", afl: AnimFrameLine{Line: "0,0, 0,0, 8, , , 1.5"}, want: true},
		{name: "Yscale", afl: AnimFrameLine{Line: "0,0, 0,0, 8, , ,,1.5"}, want: true},
		{name: "Angle", afl: AnimFrameLine{Line: "0,0, 0,0, 8, , ,, , 90"}, want: true},
		{name: "More parameters than accepted", afl: AnimFrameLine{Line: "0,0, 0,0, 8, V, AS249D200, 1.5,0.4, 90, 5"}, want: false},
	}

	for _, tt := range lineTests {

		t.Run(tt.name, func(t *testing.T) {
			got := tt.afl.IsValid()

			if got != tt.want {
				t.Errorf("got %v want %v", got, tt.want)
			}
		})

	}

}

func TestGetsAnimFrameLine(t *testing.T) {

	t.Run("Get sprite group", func(t *testing.T) {
		afl := AnimFrameLine{Line: "9000,1, 0,0, 5"}
		got := afl.GetSprGroup()
		want := int16(9000)

		if got != want {
			t.Errorf("got %v want %v", got, want)
		}
	})

	t.Run("Get sprite number", func(t *testing.T) {
		afl := AnimFrameLine{Line: "9000,1, 0,0, 5"}
		got := afl.GetSprNum()
		want := int16(1)

		if got != want {
			t.Errorf("got %v want %v", got, want)
		}
	})

	t.Run("Get pos X", func(t *testing.T) {
		afl := AnimFrameLine{Line: "9000,1, 30,15, 5"}
		got := afl.GetPosX()
		want := int16(30)

		if got != want {
			t.Errorf("got %v want %v", got, want)
		}
	})

	t.Run("Get pos Y", func(t *testing.T) {
		afl := AnimFrameLine{Line: "9000,1, 30,15, 5"}
		got := afl.GetPosY()
		want := int16(15)

		if got != want {
			t.Errorf("got %v want %v", got, want)
		}
	})

	t.Run("Get time", func(t *testing.T) {
		afl := AnimFrameLine{Line: "9000,1, 30,15, 5"}
		got := afl.GetTime()
		want := int32(5)

		if got != want {
			t.Errorf("got %v want %v", got, want)
		}
	})

	t.Run("Get default flip H when flip isn't especified", func(t *testing.T) {
		afl := AnimFrameLine{Line: "9000,1, 30,15, 5"}
		got := afl.GetFlipH()
		want := int8(1)

		if got != want {
			t.Errorf("got %v want %v", got, want)
		}
	})

	t.Run("Get flip H when is especified", func(t *testing.T) {
		afl := AnimFrameLine{Line: "9000,1, 30,15, 5, H"}
		got := afl.GetFlipH()
		want := int8(-1)

		if got != want {
			t.Errorf("got %v want %v", got, want)
		}
	})

	t.Run("Get flip H when is lowercase", func(t *testing.T) {
		afl := AnimFrameLine{Line: "9000,1, 30,15, 5, h"}
		got := afl.GetFlipH()
		want := int8(-1)

		if got != want {
			t.Errorf("got %v want %v", got, want)
		}
	})

	t.Run("Get default flip H when only flip V is specified", func(t *testing.T) {
		afl := AnimFrameLine{Line: "9000,1, 30,15, 5, v"}
		got := afl.GetFlipH()
		want := int8(1)

		if got != want {
			t.Errorf("got %v want %v", got, want)
		}
	})

	t.Run("Get flip H when both flips are specified", func(t *testing.T) {
		afl := AnimFrameLine{Line: "9000,1, 30,15, 5, HV"}
		got := afl.GetFlipH()
		want := int8(-1)

		if got != want {
			t.Errorf("got %v want %v", got, want)
		}
	})

	t.Run("Get default flip V when flip isn't especified", func(t *testing.T) {
		afl := AnimFrameLine{Line: "9000,1, 30,15, 5"}
		got := afl.GetFlipV()
		want := int8(1)

		if got != want {
			t.Errorf("got %v want %v", got, want)
		}
	})

	t.Run("Get flip V when is especified", func(t *testing.T) {
		afl := AnimFrameLine{Line: "9000,1, 30,15, 5, V"}
		got := afl.GetFlipV()
		want := int8(-1)

		if got != want {
			t.Errorf("got %v want %v", got, want)
		}
	})

	t.Run("Get flip V when is lowercase", func(t *testing.T) {
		afl := AnimFrameLine{Line: "9000,1, 30,15, 5, v"}
		got := afl.GetFlipV()
		want := int8(-1)

		if got != want {
			t.Errorf("got %v want %v", got, want)
		}
	})

	t.Run("Get default flip V when only flip H is specified", func(t *testing.T) {
		afl := AnimFrameLine{Line: "9000,1, 30,15, 5, h"}
		got := afl.GetFlipV()
		want := int8(1)

		if got != want {
			t.Errorf("got %v want %v", got, want)
		}
	})

	t.Run("Get flip V when both flips are specified", func(t *testing.T) {
		afl := AnimFrameLine{Line: "9000,1, 30,15, 5, HV"}
		got := afl.GetFlipV()
		want := int8(-1)

		if got != want {
			t.Errorf("got %v want %v", got, want)
		}
	})

}

func TestTransparencyAnimFrameLine(t *testing.T) {

	lineTests := []struct {
		name    string
		afl     AnimFrameLine
		srcWant byte
		dstWant byte
	}{
		{name: "Default on missing", afl: AnimFrameLine{Line: "9000,1, 30,15, 5"}, srcWant: byte(255), dstWant: byte(0)},
		{name: "Default on missing and scale and rotation", afl: AnimFrameLine{Line: "0,5, 0,0, 45, , , 2, , 90"}, srcWant: byte(255), dstWant: byte(0)},
		{name: "A", afl: AnimFrameLine{Line: "9000,1, 30,15, 5, , A"}, srcWant: byte(255), dstWant: byte(1)},
		{name: "A1", afl: AnimFrameLine{Line: "9000,1, 30,15, 5, , A1"}, srcWant: byte(255), dstWant: byte(128)},
		{name: "S", afl: AnimFrameLine{Line: "9000,1, 30,15, 5, , S"}, srcWant: byte(1), dstWant: byte(255)},
		{name: "Src & Dst", afl: AnimFrameLine{Line: "9000,1, 30,15, 5, , AS104D217"}, srcWant: byte(104), dstWant: byte(217)},
	}

	for _, tt := range lineTests {

		t.Run(tt.name, func(t *testing.T) {
			src, dst := tt.afl.GetTransparency()

			if src != tt.srcWant || dst != tt.dstWant {
				t.Errorf("got src: %v dst: %v | want src: %v dst: %v", src, dst, tt.srcWant, tt.dstWant)
			}
		})

	}

}

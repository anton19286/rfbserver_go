package encoder
// import "sort"

type PIXEL_T uint32
const sizeof_PIXEL_T = 4

type Pair struct {
  Key PIXEL_T
  Value int
}

type PairList []Pair

func (p PairList) Len() int { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int){ p[i], p[j] = p[j], p[i] }

type TightPalette struct {
	maxColors, numColors int

  	colorCount map[PIXEL_T]int
  	sortedList PairList
	mainColor, otherColor PIXEL_T
}

func NewTightPalette(maxColors int) TightPalette {
	tp := TightPalette{}
	tp.colorCount = make(map[PIXEL_T]int)
	tp.sortedList = make(PairList, maxColors)
	tp.maxColors = maxColors
	return tp
}

func (tp *TightPalette) Reset() {
	//TODO need fast way to reset map
	for k := range tp.colorCount {
		delete(tp.colorCount, k)
	} 
	tp.sortedList = tp.sortedList[:0]
}

// Set limit on the number of colors in the palette. Note that
// this value cannot exceed 254.
func (tp *TightPalette) SetMaxColors(maxColors int) {
	if maxColors < 0 {
		maxColors = 0
	}
	if maxColors > 254 {
		maxColors = 254
	}
	tp.maxColors = maxColors
}

// Insert new color into the palette, or increment its counter if
// the color is already there. Returns new number of colors, or
// zero if the palette is full. If the palette becomes full, it
// reports zero colors and cannot be used any more without calling
// reset().
func (tp *TightPalette) Insert(rgb PIXEL_T, numPixels int) int {
	tp.colorCount[rgb] += numPixels
	if tp.colorCount[rgb] >= tp.colorCount[tp.mainColor] {
		tp.otherColor = tp.mainColor
		tp.mainColor = rgb
	}
	tp.numColors = len(tp.colorCount)
	if len(tp.colorCount) >= tp.maxColors {
		tp.numColors = 0
	}
  	return tp.numColors
	
}

// Return the number of colors in the palette. If the palette is full,
// this function returns 0.
func (tp TightPalette) GetNumColors() int {
    return tp.numColors
}


// Return the color specified by its index in the palette.
func (tp TightPalette) GetEntry(i int) PIXEL_T {
	if i < tp.numColors {
    	return tp.sortedList[i].Key
	}
    return 0xFFFFFFFF
}

// Return the pixel counter of the color specified by its index.
func (tp TightPalette) GetCount(i int) int {
	if i < tp.numColors {
	    return tp.sortedList[i].Value
	}
	return 0
}

// Return the index of a specified color.
func (tp TightPalette) GetIndex(rgb PIXEL_T) uint8 {
    for i, c := range tp.sortedList {
      if c.Key == rgb {
        return uint8(i)
      }
    }
    return 0xFF  // no such color
}

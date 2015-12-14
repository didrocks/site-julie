package main

/* image breakpoints we handle for each resolution. We want a small one as well for images at 1x */
var IMGBREAKPOINTS = map[float32][]int{
	1:   []int{200, 480, 600, 840, 960, 1280, 1440, 1600},
	1.5: []int{200 * 1.5, 480 * 1.5, 600 * 1.5, 840 * 1.5, 960 * 1.5, 1280 * 1.5},
	2:   []int{200 * 2, 480 * 2, 6000 * 2, 840 * 2, 960 * 2},
	3:   []int{200 * 3, 480 * 3, 600 * 3, 840 * 3},
	4:   []int{200 * 4, 480 * 4, 600 * 4},
}

/* banner height at 1x */
const BANNERHEIGHT = 516

package main

import "fmt"

func outputWarnings(images [][]byte, args commonArgs) []string {
	warnings := []string{}
	for i, m := range actualOutputs(images) {
		prefix := fmt.Sprintf("output %d: ", i+1)
		if args.size != "auto" && args.size != fmt.Sprintf("%dx%d", m["width"], m["height"]) {
			warnings = append(warnings, prefix+"actual dimensions differ from requested size")
		}
		if m["format"] == "gif" {
			if m["frames"] != args.frameCount {
				warnings = append(warnings, prefix+"actual frame count differs from requested count")
			}
			if m["frames"] == 1 {
				warnings = append(warnings, prefix+"GIF has only one frame; animation requirement is not met")
			}
			if m["loop_count"] != 0 {
				warnings = append(warnings, prefix+"GIF is not marked for infinite looping")
			}
			if m["first_last_equal"] != true {
				warnings = append(warnings, prefix+"first and last composited frames differ; review the loop seam")
			}
		}
	}
	return warnings
}

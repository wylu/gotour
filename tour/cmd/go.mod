module github.com/wylu/gotour/tour/cmd

go 1.15

require (
	github.com/spf13/cobra v1.2.1
	github.com/wylu/gotour/tour/internal/word v0.0.0
)

replace github.com/wylu/gotour/tour/internal/word => ../internal/word

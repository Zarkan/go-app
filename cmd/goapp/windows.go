// +build windows

package main

import (
	"context"
)

func openCommand() string {
	return "explorer"
}

func win(ctx context.Context, args []string) {
	printErr("work in progress")
}

func mac(ctx context.Context, args []string) {
	printErr("you are not on a mac")
}

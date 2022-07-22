module github.com/esimov/caire

go 1.18

require (
	gioui.org v0.0.0-20220721082542-b67bef3e0d96
	github.com/disintegration/imaging v1.6.2
	github.com/esimov/pigo v1.4.5
	golang.org/x/exp v0.0.0-20220317015231-48e79f11773a
	golang.org/x/image v0.0.0-20220617043117-41969df76e82
	golang.org/x/term v0.0.0-20201126162022-7de9c90e9dd1
)

require (
	gioui.org/cpu v0.0.0-20220412190645-f1e9e8c3b1f7 // indirect
	gioui.org/gpu v0.0.0-00010101000000-000000000000 // indirect
	gioui.org/shader v1.0.6 // indirect
	golang.org/x/exp/shiny v0.0.0-20220713135740-79cabaa25d75 // indirect
	golang.org/x/sys v0.0.0-20220721230656-c6bc011c0c49 // indirect
	golang.org/x/text v0.3.7 // indirect
)

replace (
	gioui.org => ./vendor/gioui.org
	gioui.org/cpu => ./vendor/gioui.org/cpu
	gioui.org/gpu => ./vendor/gioui.org/gpu
)

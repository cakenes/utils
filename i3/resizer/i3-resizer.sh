#!/bin/bash
#Args:  1 grow direction, 2 shrink direction, 3 amount, 4 unit
#e.g. i3-resize.sh right left 5 px

success=$(i3-msg resize grow $1 $3 $4&> /dev/null)

if ! [[ $success == '[{"success":true}]' ]]; then
	i3-msg resize shrink $2 $3 $4&> /dev/null
fi

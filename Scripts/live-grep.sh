#!/run/current-system/sw/bin/zsh

RG_PREFIX=" g -i --line-number --column --no-heading --color=always --smart-case "
: | fzf --exact -d ':' --ansi --no-sort --disabled --query '' \
    --bind "change:reload:sleep 0.1; $RG_PREFIX {q} '$1' || true" \
    --preview "bat --color=always {1} --highlight-line {2}" \
    --preview-window 'up,60%,border-bottom,+{2}-1/3,~3' > $2

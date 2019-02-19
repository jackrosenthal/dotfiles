# Install zsh-syntax-highlighting to make this work
for b in /usr/share{/zsh/plugins,}; do
    if [[ -e $b/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh ]]; then
        source $b/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh
    fi
done

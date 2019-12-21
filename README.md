
## How to use GoLang with a private Gitlab repo? ##
Run command below
~~~~
git config --global --add url."git@gitlab.com:".insteadOf "https://gitlab.com/"
git config --global --add url."git@git.tutum.dev:".insteadOf "https://git.tutum.dev/"

export GOPRIVATE="gitlab.com/silenteer,git.tutum.dev/medi/tutum"
~~~~

Check document at 
https://www.notion.so/silenteer/b9208343cebe47cd9544beb91d045ef4?v=7a6f0c1f2b8f4732800a04270bee0b21&p=89e4dcadaa66421d9238536d358328e9
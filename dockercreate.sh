VERSION=`cat VERSION | awk -F. -v OFS=. 'NF==1{print ++$NF}; NF>1{if(length($NF+1)>length($NF))$(NF-1)++; $NF=sprintf("%0*d", length($NF), ($NF+1)%(10^length($NF))); print}'`
echo "VERSION: "$VERSION

CGO_ENABLED=0 go build && \
sudo docker build -t timz/rsshora . && \
sudo docker tag timz/rsshora timz/rsshora:$VERSION && \
sudo docker push timz/rsshora && \
echo $VERSION > VERSION



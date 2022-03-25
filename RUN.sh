go build -o opaproxy \
  opaproxy.go

if [[ $? == 0 ]]; then
    ./opaproxy -debug=3 \
    -opaport=30181 \
    -config="./config/config.yaml"
fi
echo "start ninja chatgpt mirror proxy..."

if [ -n "$HTTP_PROXY" ]; then
    echo "检测到代理HTTP_PROXY"
    #  ninjia proxies=HTTP_PROXY
    ./ninja run  -L=info --proxies="http://192.168.31.98:7860"
fi

./ninja run -D --arkose-auth-har-dir="./auth"  --arkose-gpt4-har-dir="./gpt4"  --arkose-endpoint="https://tcr9i-2.xyhelper.cn" -L=info --proxies="http://192.168.31.98:7860"
# Build the netcat image, capturing its output
BUILD_OUTPUT=$(docker build . -q -f netcat.Dockerfile -t nc)

MESSAGE="hello"
HOST=server
PORT=12345
NETWORK=tp0_testing_net

OUTPUT=$(docker run -q -it --rm --network=$NETWORK --name=netcat nc -c "echo $MESSAGE | nc $HOST $PORT" | tr -d '\r')

if [[ "$OUTPUT" == "$MESSAGE" ]]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi


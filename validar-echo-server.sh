docker build . -f netcat.Dockerfile -t nc

MESSAGE="hello"
HOST=server
PORT=12345

OUTPUT=$(docker run -it --rm --network=tp0_testing_net --name=netcat nc -c "echo $MESSAGE | nc $HOST $PORT" | tr -d '\r')

if [[ "$OUTPUT" == "$MESSAGE" ]]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi


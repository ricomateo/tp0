import socket
import signal
import logging
import sys
from common.utils import Bet

SOCKET_TIMEOUT = 5
BET_INFO_MSG_TYPE = 0
BET_CONFIRMATION_MSG_TYPE = 1

class CommunicationHandler:
    def __init__(self, port, listen_backlog):
        # Set the SIGTERM handler
        signal.signal(signal.SIGTERM, self.__sigterm_handler)
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        # Set socket timeout to check for the signal flag
        self._server_socket.settimeout(SOCKET_TIMEOUT)
        self._client_sock = None
        self._received_sig_term = False

    def accept_new_connection(self):
        """
        Accept new connections

        Function blocks for SOCKET_TIMEOUT seconds until a connection to a client is made.
        Then connection created is printed and returned.
        If a SIGTERM signal has been received by the time the socket times out,
        then the server exits gracefully.
        """

        while True:
            if self._received_sig_term:
                self.__exit_gracefully()
            try:
                # Connection arrived
                logging.info('action: accept_connections | result: in_progress')
                c, addr = self._server_socket.accept()
                logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
                self._client_sock = c
                return
            except socket.timeout:
                # This timeout allows to check for the SIGTERM signal more regularly
                continue

    def recv_msg(self):
        """
        Reads a message from the current client socket
        """
        try:
            message_type = int.from_bytes(self._client_sock.recv(1), "big")
            if message_type == BET_INFO_MSG_TYPE:
                return self.__decode_bet_info()

        except OSError as e:
            logging.error("action: receive_message | result: fail | error: {e}")

    def send_bet_confirmation(self, bet: Bet):
        """
        Sends a confirmation message for the given bet to the current socket connection.
        """
        message_type = BET_CONFIRMATION_MSG_TYPE
        document = str(bet.document).encode('utf-8')
        number = str(bet.number).encode('utf-8')
    
        # Send message type
        self._client_sock.sendall(message_type.to_bytes(1, "big"))
        # Send document's length and value
        self._client_sock.sendall(len(document).to_bytes(1, "big"))
        self._client_sock.sendall(document)
        
        # Send number's length and value
        self._client_sock.sendall(len(number).to_bytes(1, "big"))
        self._client_sock.sendall(number)
        self._client_sock.sendall(b"\n")

    def close_current_connection(self):
        self._client_sock.close()

    def __decode_bet_info(self) -> Bet:
        """
        Reads and returns a Bet message from the current socket connection.
        """
        # TODO: add error handling
        # Deserialize the fields
        agency = self.__recv_str()
        name = self.__recv_str()
        last_name = self.__recv_str()
        document = self.__recv_str()
        birthdate = self.__recv_str()
        number = self.__recv_str()
        
        logging.debug(f"Message data: agency: {agency} name: {name}, last_name: {last_name}, document: {document}, birthdate: {birthdate}, number: {number}")
        return Bet(agency, name, last_name, str(document), birthdate, str(number))
    
    def __recv_str(self) -> str:
        """
        Reads and return a string from the current socket connection.
        The string is decoded by first reading its length, and then the value.
        """
        str_len = int.from_bytes(self._client_sock.recv(1), "big")
        string = str(self._client_sock.recv(str_len), 'utf-8')
        return string

    def __sigterm_handler(self, signum, _):
        if signum == signal.SIGTERM:
            # This assignment is atomic
            self._received_sig_term = True

    def __exit_gracefully(self):
        logging.info("Exiting gracefully")
        logging.info("Shutting down server")
        self._server_socket.shutdown(socket.SHUT_RDWR)
        if self._client_sock is not None:
            logging.info("Closing socket connection")
            self._client_sock.close()
        sys.exit(0)

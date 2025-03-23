import socket
import signal
import logging
import sys
from common.utils import Bet

SOCKET_TIMEOUT = 5
BET_INFO_MSG_TYPE = 0

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
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            message_type = int.from_bytes(self._client_sock.recv(1), "big")
            if message_type == BET_INFO_MSG_TYPE:
                return self.__decode_bet_info()
            
            # TODO: Modify the send to avoid short-writes
            # self._client_sock.send("{}\n".format(msg).encode('utf-8'))
        except OSError as e:
            logging.error("action: receive_message | result: fail | error: {e}")

    def __decode_bet_info(self):
        # TODO: add error handling
        # Deserialize the name
        name_len = int.from_bytes(self._client_sock.recv(1), "big")
        name = str(self._client_sock.recv(name_len), 'utf-8')

        # Deserialize the last name
        last_name_len = int.from_bytes(self._client_sock.recv(1), "big")
        last_name = str(self._client_sock.recv(last_name_len), 'utf-8')
        
        # Deserialize the document
        document_len = int.from_bytes(self._client_sock.recv(1), "big")
        document = str(self._client_sock.recv(document_len), 'utf-8')
        
        # Deserialize the date of birth
        birthdate_len = int.from_bytes(self._client_sock.recv(1), "big")
        birthdate = str(self._client_sock.recv(birthdate_len), 'utf-8')

        # Deserialize the number
        number_len = int.from_bytes(self._client_sock.recv(1), "big")
        number = str(self._client_sock.recv(number_len), 'utf-8')

        # TODO: include agency in the message
        agency = "1"
        
        logging.info(f"Message data: name: {name}, last_name: {last_name}, document: {document}, birthdate: {birthdate}, number: {number}")
        return Bet(agency, name, last_name, str(document), birthdate, str(number))
    

    def close_current_connection(self):
        self._client_sock.close()

    
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

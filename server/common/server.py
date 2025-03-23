import logging
from common.communication import CommunicationHandler

SOCKET_TIMEOUT = 5
BET_INFO_MSG_TYPE = 0

class Server:
    def __init__(self, port, listen_backlog):
        self.__communication_handler = CommunicationHandler(port, listen_backlog)
        

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        while True:
            self.__communication_handler.accept_new_connection()
            msg = self.__communication_handler.recv_msg()
            # TODO: process the message
            logging.info(f"Received msg: {msg}")
            self.__communication_handler.close_current_connection()

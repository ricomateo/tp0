import logging
from common.communication import CommunicationHandler
from common.error import MessageReceptionError
from common.utils import store_bets, Bet

SOCKET_TIMEOUT = 5
BET_INFO_MSG_TYPE = 0

class Server:
    def __init__(self, port, listen_backlog):
        self.__communication_handler = CommunicationHandler(port, listen_backlog)
        

    def run(self):
        """
        Dummy Server loop

        Server that accept new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        while True:
            try:
                self.__communication_handler.accept_new_connection()
                bets = self.__communication_handler.recv_msg()
                self._handle_batch_message(bets)

            except Exception as e:
                logging.error(f"failed to handle client connection. Error: {e}")
                
            finally:
                self.__communication_handler.close_current_connection()

    def _handle_batch_message(self, batch: list[Bet]):
        try:
            bets = batch
            store_bets(bets)
            logging.info(f"action: apuesta_recibida | result: success | cantidad: {len(bets)}")
            self.__communication_handler.send_batch_success()
        except Exception as e:
            logging.error(f"action: apuesta_recibida | result: fail | cantidad: {len(bets)}")
            self.__communication_handler.send_batch_failure()

import logging
from common.communication import *
from common.error import MessageReceptionError
from common.utils import store_bets


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
                # TODO: consider creating a Message class
                message_type, payload = self.__communication_handler.recv_msg()
                if message_type == BET_BATCH_MSG_TYPE:
                    bets = payload
                    store_bets(bets)
                    logging.info(f"action: apuesta_recibida | result: success | cantidad: {len(bets)}")
                    self.__communication_handler.send_batch_success()

                elif message_type == FINALIZATION_MSG_TYPE:
                    agency_id = payload
                    logging.info(f"Agency with id {agency_id} finished!!!")

            except MessageReceptionError as e:
                self.__communication_handler.send_batch_failure()

            except Exception as e:
                logging.error(f"failed to handle client connection. Error: {e}")
                
            finally:
                self.__communication_handler.close_current_connection()


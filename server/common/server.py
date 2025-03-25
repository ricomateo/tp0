import logging
from common.communication import *
from common.error import MessageReceptionError
from common.utils import *


class Server:
    def __init__(self, port, listen_backlog, number_of_clients):
        self.__communication_handler = CommunicationHandler(port, listen_backlog)
        self.finished_agencies = set()
        self.winners_by_agency = {}
        self.number_of_clients = int(number_of_clients)

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
                    self._set_agency_as_finished(agency_id)
                    if self._all_agencies_finished():
                        logging.info(f"action: sorteo | result: success")
                        self._load_winners()

                elif message_type == GET_WINNERS_MSG_TYPE:
                    agency_id = payload
                    if self._all_agencies_finished():
                        winners = self.winners_by_agency.get(agency_id, [])
                        self.__communication_handler.send_winners(winners)
                    else:
                        self.__communication_handler.send_no_winners_yet()
                else:
                    # TODO: raise an error
                    logging.info(f"Invalid message_type = {message_type}")
            except MessageReceptionError as e:
                self.__communication_handler.send_batch_failure()

            except Exception as e:
                logging.error(f"failed to handle client connection. Error: {e}")
                
            finally:
                self.__communication_handler.close_current_connection()

    def _all_agencies_finished(self) -> bool:
        agencies_ids = list(range(1, self.number_of_clients + 1))
        for id in agencies_ids:
            if id not in self.finished_agencies:
                return False
        return True
    
    def _set_agency_as_finished(self, agency: int):
        self.finished_agencies.add(agency)

    def _load_winners(self):
        bets = load_bets()
        for bet in bets:
            if has_won(bet):
                self.winners_by_agency.setdefault(bet.agency, []).append(bet.document)

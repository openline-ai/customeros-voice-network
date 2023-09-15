import psycopg2


class KamailioDatabase:
    connection = None
    conn_string = None

    def __init__(self, host: str, database: str, user: str, password: str):
        self.conn_string = "host='%s' dbname='%s' user='%s' password='%s'" % (host, database, user, password)
        self.connection = psycopg2.connect(self.conn_string)
        self.connection.set_session(autocommit=True)

    def test_connection(self):
        try:
            cur = self.connection.cursor()
            cur.execute('SELECT 1')
        except psycopg2.OperationalError:
            self.connection = psycopg2.connect(self.conn_string)
            self.connection.set_session(autocommit=True)

    def lookup_carrier(self, carrier: str):
        self.test_connection()
        with self.connection.cursor() as cur:

            cur.execute("SELECT username, ha1, domain FROM openline_carrier WHERE carrier_name=%s",
                        (carrier,))
            record = cur.fetchone()
            if record is not None:
                return {"username": record[0],
                        "ha1": record[1],
                        "domain": record[2]}
        return None

    def find_sipuri_mapping(self, sipuri: str):
        self.test_connection()
        with self.connection.cursor() as cur:

            cur.execute("SELECT e164, alias, carrier_name, sipuri, profile_id "
                        + "FROM openline_number_mapping WHERE sipuri=%s OR phoneuri=%s",
                        (sipuri, sipuri))
            record = cur.fetchone()
            if record is not None:
                return {"e164": record[0],
                        "alias": record[1],
                        "carrier": record[2],
                        "sipuri": record[3],
                        "profile_id": record[4]
                        }
        return None

    def find_e164_mapping(self, e164: str, carrier: str):
        self.test_connection()
        with self.connection.cursor() as cur:
            cur.execute("SELECT onm.sipuri, onm.phoneuri, onm.profile_id, onm.voicemail_id, of.enabled, of.e164 "
                        + "FROM openline_number_mapping onm LEFT JOIN openline_forwarding of ON (onm.forwarding_id=of.id) "
                        + "WHERE onm.e164=%s AND onm.carrier_name=%s", (e164, carrier))
            record = cur.fetchone()
            if record is not None:
                return {"sipuri": record[0],
                        "phoneuri": record[1],
                        "profile_id": record[2],
                        "voicemail_id": record[3],
                        "forwarding_enabled": record[4],
                        "forwarding_e164": record[5]}
        return None

    def find_profile(self, profile_id: int):
        if profile_id is None:
            return None

        self.test_connection()
        with self.connection.cursor() as cur:
            cur.execute("SELECT profile_name, call_webhook, recording_webhook, api_key "
                        + "FROM openline_profile WHERE id=%s", (profile_id,))
            record = cur.fetchone()
            if record is not None:
                return {"profile_name": record[0],
                        "call_webhook": record[1],
                        "recording_webhook": record[2],
                        "api_key": record[3]
                        }

    def find_voicemail(self, voicemail_id: int):
        if voicemail_id is None:
            return None

        self.test_connection()
        with self.connection.cursor() as cur:
            cur.execute("SELECT object_id, timeout "
                        + "FROM openline_voicemail WHERE id=%s", (voicemail_id,))
            record = cur.fetchone()
            if record is not None:
                return {"prompt_object_id": record[0],
                        "timeout": record[1]
                        }
        return None



class TestKamailioDatabase:
    _mock = {}

    def __init__(self):
        pass

    def lookup_carrier(self, carrier:str):
        return self._mock['lookup_carrier'](carrier)

    def find_sipuri_mapping(self, sipuri:str):
        return self._mock['find_sipuri_mapping'](sipuri)

    def find_e164_mapping(self, e164:str, carrier:str):
        return self._mock['find_e164_mapping'](e164, carrier)

    def find_profile(self, profile_id: int):
        if profile_id is None:
            return None
        return self._mock['find_profile'](profile_id)

    def find_voicemail(self, voicemail_id: int):
        if voicemail_id is None:
            return None
        return self._mock['find_voicemail'](voicemail_id)

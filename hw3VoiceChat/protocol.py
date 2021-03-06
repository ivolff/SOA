import enum

class DataType(enum.Enum):
    ClientData = 1
    Handshake = 2
    Disconnetion = 3
    NowConnected = 4

class Protocol:
    CLIENT_DATA_MIN = 0
    CLIENT_DATA_MAX = 50
    HANDSHAKE = 51
    DISCONNECTION = 52
    NOW_CONNECTED = 53

    typeToOrd = {DataType.ClientData:CLIENT_DATA_MIN, DataType.Handshake:HANDSHAKE, DataType.Disconnetion:DISCONNECTION, DataType.NowConnected:NOW_CONNECTED}
    ordToType = {v: k for k, v in typeToOrd.items()}

    def __init__(self, dataType=None, head=None, data=None, datapacket=None):
        if dataType is not None:
            self.head = Protocol.typeToOrd[dataType]
        else:
            self.head = datapacket[0] if head is None else head
        self.data = datapacket[1:] if data is None else data
        self.DataType = Protocol.getDataType(self.head)

    @staticmethod
    def getDataType(head):
        if head <= Protocol.CLIENT_DATA_MAX and head >= Protocol.CLIENT_DATA_MIN:
            return DataType.ClientData
        try:
            return Protocol.ordToType[head]
        except:
            return None

    def out(self):
        bytearr = bytearray(b'')
        bytearr.append(self.head)
        return bytes(bytearr + self.data)
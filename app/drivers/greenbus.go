package drivers

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/wdreeveii/termioslib"
	"io/ioutil"
	"log"
	"os"
	"time"
	"github.com/wdreeveii/timecl/app/logger"
	"github.com/wdreeveii/timecl/app/network_manager"
)

//ioutil.Discard
//os.Stderr
var LOG = log.New(os.Stderr, "GreenBus ", log.Ldate|log.Ltime)
var DEBUG = log.New(ioutil.Discard, "GreenBus ", log.Ldate|log.Ltime)

type GreenBus struct {
	StopChan   chan bool
	List_ports chan chan []network_manager.BusDef
}

func (d GreenBus) Init(port string, network_id int) {
	d.StopChan = make(chan bool)
	go d.runmaster(port, network_id)
}

func (d GreenBus) Stop() {
	d.StopChan <- true
}

func (d GreenBus) ListPorts() (result []network_manager.BusDef) {
	if d.List_ports != nil {
		var res = make(chan []network_manager.BusDef)
		d.List_ports <- res
		result = <-res
	}
	return
}

type Cmd_t struct {
	Payload      []byte
	Mtype        uint8
	Rtype        uint8
	ReplyHandler func(Message_t)
	CmdInfo      interface{}
}

type Port_t struct {
	Type  network_manager.PortFunction
	Value uint16
}

type GreenBusDevice struct {
	Addr      uint8
	Mac       uint32
	ModelName string
	Serial    string
	// things like number of errors
	Cmds  []Cmd_t
	Ports []Port_t
}

func (d GreenBusDevice) String() string {
	return fmt.Sprintf("<Addr: %v, Mac: %v, Model: %v>", d.Addr, d.Mac, d.ModelName)
}

type DeviceList []GreenBusDevice

func (d DeviceList) Find(Mac uint32) (int, bool) {
	for idx, val := range d {
		if val.Mac == Mac {
			return idx, true
		}
	}
	return 0, false
}

type portval_payload []byte
type portval_struct struct {
	Id    uint8
	Value uint16
}

func (def portval_payload) ToStruct() (dst portval_struct) {
	dst.Id = def[0]
	dst.Value = binary.LittleEndian.Uint16(def[1:])
	return
}

type values_payload []byte

func (def values_payload) ToStruct() (dst []portval_struct) {
	if len(def) > 0 {
		num_ports := uint8(def[0])
		if len(def) >= int(1+(3*num_ports)) {
			for i := 0; i < int(num_ports); i++ {
				dst = append(dst, portval_payload(def[1+(i*3):4+(i*3)]).ToStruct())
			}
		}
	}
	return
}

type portdef_payload []byte
type portdef_struct struct {
	Type  uint8
	Index uint8
}

func (def portdef_payload) ToStruct() (dst portdef_struct) {
	dst.Type = def[0]
	dst.Index = def[1]
	return
}

type interrogate_payload []byte
type interrogate_reply struct {
	Model  string
	Serial string
	Ports  []portdef_struct
}

func (payload interrogate_payload) ToStruct() (dst interrogate_reply) {
	if len(payload) > 0 {
		model_len := uint8(payload[0])
		if len(payload) >= 1+int(model_len) {
			dst.Model = string(payload[1 : 1+model_len])
			if len(payload) > 1+int(model_len) {
				serial_len := uint8(payload[1+model_len])
				if len(payload) >= 2+int(model_len)+int(serial_len) {
					dst.Serial = string(payload[model_len+2 : model_len+2+serial_len])
					ports_offset := 2 + model_len + serial_len
					if len(payload) > 2+int(model_len)+int(serial_len) {
						num_ports := uint8(payload[ports_offset])
						if len(payload) == int(ports_offset+1+(2*num_ports)) {
							for ii := uint8(0); ii < uint8(num_ports); ii++ {
								portdef_start := ports_offset + 1 + (ii * 2)
								dst.Ports = append(dst.Ports, portdef_payload(payload[portdef_start:portdef_start+2]).ToStruct())
							}
						}
					}
				}
			}
		}
	}
	return
}

func (structure interrogate_reply) ToByte() (payload interrogate_payload) {
	return
}

const (
	PORT_OUTPUT  uint8 = 0
	PORT_BINPUT  uint8 = 2
	PORT_AINPUT  uint8 = 4
	PORT_AOUTPUT uint8 = 8

	FIND_DEVICES uint8 = 72
	NEED_ADDR    uint8 = 73
	ACK_DEVICE   uint8 = 74
	ACK_REPLY    uint8 = 75

	INTERROGATE       uint8 = 80
	INTERROGATE_REPLY uint8 = 81

	PING       uint8 = 44
	PING_REPLY uint8 = 45

	RESET_NETWORK uint8 = 58

	GET       uint8 = 36
	GET_REPLY uint8 = 37
	SET       uint8 = 38
	SET_REPLY uint8 = 39
)

type Header_t struct {
	Destination uint8
	Mtype       uint8
	Length      uint16
	Mac         uint32
	Crc         uint16
}

type Message_t struct {
	Header  Header_t
	Payload []byte
}

func crc16_update(crc uint16, a byte) uint16 {
	var i int
	crc ^= uint16(a)
	for i = 0; i < 8; i++ {
		if crc&1 != 0 {
			crc = (crc >> 1) ^ 0xA001
		} else {
			crc = (crc >> 1)
		}
	}
	return crc
}

func ToByte(msg Message_t) []byte {
	b := new(bytes.Buffer)
	b.Write([]byte("A"))
	binary.Write(b, binary.LittleEndian, msg.Header)
	b.Write([]byte("A"))
	b.Write(msg.Payload)
	bytes := b.Bytes()
	var checksum uint16 = 0xffff
	for i := 0; i < len(bytes); i++ {
		checksum = crc16_update(checksum, bytes[i])
	}
	binary.Write(b, binary.LittleEndian, checksum)
	return b.Bytes()
}

func HeaderCRC(msg *Message_t) uint16 {
	msg.Header.Length = uint16(len(msg.Payload) + 2) // add space for the crc
	b := new(bytes.Buffer)
	b.Write([]byte("A"))
	binary.Write(b, binary.LittleEndian, msg.Header)
	bytes := b.Bytes()
	var checksum uint16 = 0xffff
	for i := 0; i < len(bytes)-2; i++ {
		checksum = crc16_update(checksum, bytes[i])
	}
	return checksum
}

func setup_serial(ser *os.File) (*termioslib.Termios, error) {
	var (
		err          error
		orig_termios *termioslib.Termios
		work_termios *termioslib.Termios
	)
	orig_termios = new(termioslib.Termios)
	work_termios = new(termioslib.Termios)
	// flush all buffers
	if err = termioslib.Flush(ser.Fd(), termioslib.TCIOFLUSH); err != nil {
		return orig_termios, err
	}

	// save a copy of the original terminal configuration
	if err = termioslib.Getattr(ser.Fd(), orig_termios); err != nil {
		return orig_termios, err
	}

	// get a working copy
	if err = termioslib.Getattr(ser.Fd(), work_termios); err != nil {
		return orig_termios, err
	}

	work_termios.C_iflag &= ^(termioslib.BRKINT | termioslib.INPCK | termioslib.ISTRIP | termioslib.ICRNL | termioslib.IXON)
	work_termios.C_oflag &= ^(termioslib.OPOST | termioslib.ONLCR)
	work_termios.C_lflag &= ^(termioslib.ISIG | termioslib.ICANON | termioslib.IEXTEN | termioslib.ECHO | termioslib.ECHOE | termioslib.ECHOK)
	//work_termios.C_cflag &= ^(termioslib.CSIZE | termioslib.PARENB | termioslib.PARODD | termioslib.HUPCL | termioslib.CSTOPB )
	work_termios.C_cflag |= (termioslib.CS8)
	work_termios.C_cc[termioslib.VMIN] = 1
	work_termios.C_cc[termioslib.VTIME] = 0

	if err = termioslib.Setispeed(work_termios, termioslib.B57600); err != nil {
		return orig_termios, err
	}
	if err = termioslib.Setospeed(work_termios, termioslib.B57600); err != nil {
		return orig_termios, err
	}

	// set the working copy
	if err = termioslib.Setattr(ser.Fd(), termioslib.TCSANOW, work_termios); err != nil {
		return orig_termios, err
	}
	return orig_termios, err
}

func SendDeviceCmd(w chan Message_t, addr uint8, mac uint32, cmd Cmd_t) {
	var msg Message_t
	msg.Header.Destination = addr
	msg.Header.Mac = mac
	msg.Payload = cmd.Payload
	msg.Header.Mtype = cmd.Mtype

	w <- msg
	DEBUG.Println("beat")
}

type ReplyInfo struct {
	ReplyPkt Message_t
	Cmd      Cmd_t
}

func PingDevice(r chan Message_t, w chan Message_t, ping_reply chan ReplyInfo, addr uint8, mac uint32, cmd Cmd_t) {
	SendDeviceCmd(w, addr, mac, cmd)
	DEBUG.Println("Ping wait ", addr)
	var result ReplyInfo
	result.Cmd = cmd
	select {
	case rmsg := <-r:
		result.ReplyPkt = rmsg
		// else report protocol error
		DEBUG.Printf("ping read: %#v\n", rmsg)
		//time.Sleep(1 * time.Second)
	case <-time.After(200 * time.Millisecond):
		// report timeout error
		DEBUG.Printf("Ping read timeout\n")
	}
	ping_reply <- result
}

func SendFindDevices(w chan Message_t) {
	var msg Message_t
	msg.Payload = []byte("HAHAHA")

	msg.Header.Destination = 0
	msg.Header.Mtype = FIND_DEVICES
	msg.Header.Mac = 0

	w <- msg
	DEBUG.Println("Looking for new devices..")
}

func FindDevices(r chan Message_t, w chan Message_t, ack_new_devices chan []uint32) {
	SendFindDevices(w)
	var new_devices []uint32
	timeout := time.After(200 * time.Millisecond)
Loop:
	for {
		select {
		case <-timeout:
			DEBUG.Println("Find Devices Timeout")
			break Loop
		case rmsg := <-r:
			DEBUG.Println("Find Pkt Recv")
			if rmsg.Header.Mtype == NEED_ADDR { // and payload == "NEED ADDR"
				DEBUG.Println("Found A Device!")
				new_devices = append(new_devices, rmsg.Header.Mac)
			}
		}
	}
	ack_new_devices <- new_devices
	DEBUG.Println("Finish Find Devices")
}

func SendDeviceAck(w chan Message_t, device_addr uint8, mac uint32) {
	var msg Message_t

	msg.Payload = []byte(fmt.Sprintf("%d", device_addr))
	msg.Header.Destination = 0
	msg.Header.Mtype = ACK_DEVICE
	msg.Header.Mac = mac

	w <- msg
}

type AckReplyInfo struct {
	ReplyPkt    Message_t
	device_addr uint8
	device_mac  uint32
}

func AckDevice(r chan Message_t, w chan Message_t, ack_reply chan AckReplyInfo, device_addr uint8, mac uint32) {
	SendDeviceAck(w, device_addr, mac)
	var msg AckReplyInfo
	msg.device_addr = device_addr
	msg.device_mac = mac

	select {
	case rmsg := <-r:
		msg.ReplyPkt = rmsg
	case <-time.After(40 * time.Millisecond):
		DEBUG.Println("Address assignment response not received.")
	}
	ack_reply <- msg
}

func (d *GreenBusDevice) getNextCmd(devices DeviceList) Cmd_t {
	var cmd Cmd_t
	if len(d.Cmds) > 0 {
		cmd = d.Cmds[0]
		d.Cmds = d.Cmds[1:]
	} else {
		cmd = Cmd_t{Payload: []byte("PING"),
			Mtype: PING,
			Rtype: PING_REPLY,
			ReplyHandler: func(msg Message_t) {
				data := values_payload(msg.Payload).ToStruct()
				for _, v := range data {
					if int(d.Addr) < len(devices) && int(v.Id) < len(devices[d.Addr].Ports) {
						devices[d.Addr].Ports[v.Id].Value = v.Value
					}
				}
			}}
	}
	return cmd
}

func (d *GreenBus) runmaster(port string, network_id int) {
	var (
		err          error
		orig_termios *termioslib.Termios
		ser          *os.File
	)

	defer func() {
		LOG.Println("Serial Exiting...")
		if err != nil {
			logger.PublishOneError(fmt.Errorf("Error occurred and driver shutting down:", err))
		}
	}()

	ser, err = os.OpenFile(port, os.O_RDWR, 777)
	if err != nil {
		return
	}

	defer func() {
		LOG.Println("Closing serial port")
		ser.Close()
	}()

	orig_termios, err = setup_serial(ser)

	if err != nil {
		return
	}

	defer func() {
		// set the settings back to the original when the program exits
		LOG.Println("Resetting termios")
		err = termioslib.Setattr(ser.Fd(), termioslib.TCSANOW, orig_termios)
	}()

	r := make(chan Message_t)
	w := make(chan Message_t)

	defer func() {
		close(r)
		close(w)
	}()

	defer func() {
		LOG.Println("Greenbus stopping!!")
	}()

	d.List_ports = make(chan (chan []network_manager.BusDef))

	go WriteMessages(ser, w)
	go ReadMessages(ser, r)
	LOG.Println("Resetting Client Addrs")
	w <- Message_t{Payload: []byte("RESET ADDR"), Header: Header_t{Destination: 0, Mtype: RESET_NETWORK, Mac: 0}}

	var devices DeviceList = []GreenBusDevice{GreenBusDevice{Addr: 0, Mac: 0}, GreenBusDevice{Addr: 1, Mac: 0}}
	var ping_reply chan ReplyInfo
	var ack_new_devices chan []uint32
	var ack_reply chan AckReplyInfo
	var next_device int
	next_device = 0
	var state_ping_devices bool
	state_ping_devices = true
	var state_find_devices bool
	state_find_devices = false
	var new_devices []uint32

	network_subscription := network_manager.SubscribeNetwork(network_id)
	defer network_subscription.Cancel()

	for {
		// are we pinging devices?
		if state_ping_devices {
			DEBUG.Println("State Ping")
			// are we waiting for a reply? - no
			if ping_reply == nil {
				DEBUG.Println("State Ping Reply")
				device_len := len(devices)
				if device_len < 2 {
					device_len = 2
				}
				if next_device == device_len-2 { // we have reached the end of devices to ping
					next_device = 0
					state_ping_devices = false
					state_find_devices = true
					ack_new_devices = make(chan []uint32)
					go FindDevices(r, w, ack_new_devices)
				} else { // we have a device to ping
					ping_reply = make(chan ReplyInfo, 1)
					DEBUG.Println("Cmds: ", devices[next_device+2].Cmds)
					var cmd Cmd_t
					cmd = devices[next_device+2].getNextCmd(devices)

					go PingDevice(r, w, ping_reply, devices[next_device+2].Addr, devices[next_device+2].Mac, cmd)
					DEBUG.Println("Cmds2: ", devices[next_device+2].Cmds)
					next_device += 1
				}
			}
		}
		if state_find_devices {
			DEBUG.Println("State Find")
			//are we waiting for new devices? - no
			if ack_new_devices == nil {
				// are we waiting for an ack reply? - no
				if ack_reply == nil {
					DEBUG.Println("State Find Ack")
					DEBUG.Println("New Devices ", new_devices)
					if next_device == len(new_devices) { // we have reached the end of devices to ack
						new_devices = []uint32{}
						next_device = 0
						state_ping_devices = true
						state_find_devices = false
						device_len := len(devices)
						if device_len < 2 {
							device_len = 2
						}
						if next_device == device_len-2 {
							next_device = 0
							state_ping_devices = false
							state_find_devices = true
							ack_new_devices = make(chan []uint32)
							go FindDevices(r, w, ack_new_devices)
						} else {
							ping_reply = make(chan ReplyInfo, 1)
							DEBUG.Println("Cmds: ", devices[next_device+2].Cmds)
							var cmd Cmd_t
							cmd = devices[next_device+2].getNextCmd(devices)

							go PingDevice(r, w, ping_reply, devices[next_device+2].Addr, devices[next_device+2].Mac, cmd)
							DEBUG.Println("Cmds2: ", devices[next_device+2].Cmds)
							next_device += 1
						}
					} else {
						device_mac := new_devices[next_device]
						next_device += 1
						ack_reply = make(chan AckReplyInfo, 1)
						var device_addr uint8
						device_idx, device_in_list := devices.Find(device_mac)
						if !device_in_list {
							device_addr = uint8(len(devices))
						} else {
							device_addr = uint8(device_idx)
						}
						go AckDevice(r, w, ack_reply, device_addr, device_mac)
					}
				}
			}
		}
		select {
		case req := <-d.List_ports:
			var res = make([]network_manager.BusDef, 0)
			res = append(res, network_manager.BusDef{BusID: 0})
			res[0].DeviceList = make([]network_manager.DeviceDef, 0)
			for _, device := range devices[2:] {
				var dev network_manager.DeviceDef
				dev.DeviceID = device.Mac
				var ports = make([]network_manager.PortDef, 0)
				for idx, port := range device.Ports {
					ports = append(ports, network_manager.PortDef{PortID: uint32(idx), Type: port.Type})
				}
				dev.PortList = ports
				res[0].DeviceList = append(res[0].DeviceList, dev)
			}
			req <- res
		case event := <-network_subscription.New:
			//fmt.Println("Driver event", event.Type)
			switch {
			case event.Type == "set":
				cmd := event.Data.([]network_manager.SetData)
				var ports_on_valid_devices = make(map[int][]network_manager.SetData)
				for _, v := range cmd {
					device_idx, device_in_list := devices.Find(uint32(v.DeviceID))
					if device_in_list {
						ports_on_valid_devices[device_idx] = append(ports_on_valid_devices[device_idx], v)
					}
				}
				tmpval := make([]byte, 2, 2)
				for k, ports_on_one_device := range ports_on_valid_devices {
					var devicekey = k
					if len(ports_on_one_device) > 0 {
						payload := []byte{uint8(len(ports_on_one_device))}
						for _, v := range ports_on_one_device {
							binary.LittleEndian.PutUint16(tmpval, uint16(v.Value))
							payload = append(payload, uint8(v.PortID), tmpval[0], tmpval[1])
						}
						set_cmd := []Cmd_t{Cmd_t{Payload: payload,
							Mtype: SET,
							Rtype: SET_REPLY,
							ReplyHandler: func(msg Message_t) {
								data := values_payload(msg.Payload).ToStruct()
								for _, v := range data {
									if devicekey < len(devices) && int(v.Id) < len(devices[devicekey].Ports) {
										devices[devicekey].Ports[v.Id].Value = v.Value
									}
								}
							}}}
						devices[k].Cmds = append(set_cmd, devices[k].Cmds...)
					}
				}
				// set a port
			case event.Type == "get":
				cmd, ok := event.Data.(network_manager.GetData)
				if !ok {
					logger.PublishOneError(fmt.Errorf("Greenbus: Improperly formatted get port request."))
				} else {
					sent := false
					device_idx, found := devices.Find(uint32(cmd.DeviceID))
					if found {
						if cmd.PortID < len(devices[device_idx].Ports) {
							sent = true
							cmd.Recv <- float64(devices[device_idx].Ports[cmd.PortID].Value)
						}
					}
					if !sent {
						close(cmd.Recv)
					}
				}
			}
		case reply := <-ping_reply:
			DEBUG.Println("Do Ping Reply")
			ping_reply = nil
			if reply.ReplyPkt.Header.Mtype == reply.Cmd.Rtype {
				if reply.Cmd.ReplyHandler != nil {
					reply.Cmd.ReplyHandler(reply.ReplyPkt)
				}
			}
		case msg_devices := <-ack_new_devices:
			DEBUG.Println("Do Ack Devices")
			ack_new_devices = nil
			new_devices = msg_devices
		case rmsg := <-ack_reply:
			DEBUG.Println("Do Ack Reply")
			ack_reply = nil
			if rmsg.ReplyPkt.Header.Mtype == ACK_REPLY && rmsg.ReplyPkt.Header.Mac == rmsg.device_mac { // and payload == "ACK"
				interrogate_cmd := []Cmd_t{Cmd_t{Payload: []byte("INTERROGATE"),
					Mtype: INTERROGATE,
					Rtype: INTERROGATE_REPLY,
					ReplyHandler: func(msg Message_t) {
						DEBUG.Println("INTERROGATE REPLY")
						dev_properties := interrogate_payload(msg.Payload).ToStruct()
						devices[rmsg.device_addr].ModelName = dev_properties.Model
						devices[rmsg.device_addr].Serial = dev_properties.Serial

						LOG.Println("model: ", devices[rmsg.device_addr].ModelName)
						LOG.Println("Serial: ", dev_properties.Serial)
						// parse the msg and init the ports
						var ports []Port_t
						LOG.Println("rports:", dev_properties.Ports)
						for _, val := range dev_properties.Ports {
							DEBUG.Println("SETTING UP PORT")
							var port Port_t
							switch {
							case val.Type == PORT_OUTPUT:
								port.Type = network_manager.BOutput
							case val.Type == PORT_BINPUT:
								port.Type = network_manager.BInput
							case val.Type == PORT_AINPUT:
								port.Type = network_manager.AInput
							case val.Type == PORT_AOUTPUT:
								port.Type = network_manager.AOutput
							}
							ports = append(ports, port)
						}
						DEBUG.Println("PORTS:", ports)
						devices[rmsg.device_addr].Ports = ports
						network_manager.Publish(network_manager.NewEvent(network_id, "port_change", nil))
					}}}
				device_idx, device_in_list := devices.Find(rmsg.device_mac)
				if !device_in_list {
					new_device := GreenBusDevice{Addr: uint8(len(devices)),
						Mac:  rmsg.device_mac,
						Cmds: interrogate_cmd}
					devices = append(devices, new_device)
				} else {
					devices[device_idx].Cmds = append(interrogate_cmd, devices[device_idx].Cmds...)
				}
			}
		}
		DEBUG.Println("DEVICES: ", devices)
	}
}

func ReadMessages(c *os.File, r chan Message_t) {
	var buffer []byte
	var hsize int = binary.Size(Header_t{})
	in := make([]byte, 1)
	for {
		num_read, err := c.Read(in)
		if err != nil {
			LOG.Println(err)
			return
		}
		if num_read > 0 {
			buffer = append(buffer, in[0])
			var header Header_t
			if len(buffer) == hsize+2 {
				if buffer[0] != 'A' || buffer[hsize+1] != 'A' {
					buffer = buffer[1:]
					continue
				}
				var checksum uint16 = 0xffff
				for i := 0; i < hsize-1; i++ {
					checksum = crc16_update(checksum, buffer[i])
				}
				iobuffer := new(bytes.Buffer)
				iobuffer.Write(buffer[1 : hsize+1])
				err = binary.Read(iobuffer, binary.LittleEndian, &header)
				if err != nil {
					LOG.Println(err)
				}

				if checksum != header.Crc {
					logger.PublishOneError(fmt.Errorf("Greenbus: Header CRC doesn't match"))
					buffer = buffer[1:]
					continue
				}
			}
			if len(buffer) > hsize+2 {
				iobuffer := new(bytes.Buffer)
				iobuffer.Write(buffer[1 : hsize+1])
				err = binary.Read(iobuffer, binary.LittleEndian, &header)
				if err != nil {
					LOG.Println(err)
				}
				if len(buffer) == hsize+2+int(header.Length) {
					var checksum uint16 = 0xffff
					for i := 0; i < len(buffer)-2; i++ {
						checksum = crc16_update(checksum, buffer[i])
					}
					if checksum == binary.LittleEndian.Uint16(buffer[len(buffer)-2:]) {
						var msg Message_t
						msg.Header = header
						msg.Payload = buffer[hsize+2 : len(buffer)-2]
						r <- msg
					}
					buffer = make([]byte, 0)
				}
				if len(buffer) > hsize+2+int(header.Length) {
					buffer = make([]byte, 0)
				}
			}

		}
	}
}
func WriteMessages(c *os.File, w chan Message_t) {
	var msg Message_t
	var ok bool
	for {
		msg, ok = <-w
		if !ok {
			return
		}
		msg.Header.Crc = HeaderCRC(&msg)
		DEBUG.Printf("Msg To: %d Type: %d Length: %d\n", msg.Header.Destination, msg.Header.Mtype, msg.Header.Length)
		_, err := c.Write(ToByte(msg))
		if err != nil {
			LOG.Println(err)
			//return
		}
	}
}

func init() {
	network_manager.RegisterDriver("greenbus", GreenBus{})
}

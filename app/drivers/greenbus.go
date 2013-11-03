package drivers

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"time"
	//"github.com/robfig/revel"
	//"github.com/mewkiz/pkg/hashutil/crc16"
	//"github.com/robfig/revel/modules/jobs/app/jobs"
	"github.com/wdreeveii/termioslib"
	"timecl/app/network_manager"
)

type GreenBus struct {
	StopChan chan bool
	GetChan  chan network_manager.GetDrvCmd
	SetChan  chan network_manager.SetDrvCmd
	SendCmd  chan Cmd_t
}

func (d GreenBus) Init(port string) {
	fmt.Println("CHANNEL: ", d.StopChan)
	go d.runmaster(port)
}

func (d GreenBus) Stop() {
	d.StopChan <- true
}

func (d GreenBus) Copy() network_manager.DriverInterface {
	var b GreenBus
	b = d
	b.StopChan = make(chan bool)
	b.GetChan = make(chan network_manager.GetDrvCmd)
	b.SetChan = make(chan network_manager.SetDrvCmd)
	b.SendCmd = make(chan Cmd_t)
	return b
}

func (d GreenBus) Get(cmd network_manager.GetDrvCmd) {
	d.GetChan <- cmd
}

func (d GreenBus) Set(cmd network_manager.SetDrvCmd) {
	d.SetChan <- cmd
}

func (d GreenBus) GetBusList() map[int]string {
	return map[int]string{0: "Main"}
}

func (d GreenBus) GetDeviceList(bus int) map[int]string {
	return map[int]string{}
}

type Cmd_t struct {
	Payload      []byte
	Mtype        uint8
	Rtype        uint8
	ReplyHandler func(Message_t)
	CmdInfo      interface{}
}

type PortFunction int

const (
	Input PortFunction = iota
	Output
)

type Port_t struct {
	Type  PortFunction
	Value interface{}
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
	return fmt.Sprintf("<Addr: %v, Mac: %v>", d.Addr, d.Mac)
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
						if len(payload) > int(ports_offset+1+(2*num_ports)) {
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
	fmt.Println("beat")
}

type ReplyInfo struct {
	ReplyPkt Message_t
	Cmd      Cmd_t
}

func PingDevice(r chan Message_t, w chan Message_t, ping_reply chan ReplyInfo, addr uint8, mac uint32, cmd Cmd_t) {
	SendDeviceCmd(w, addr, mac, cmd)
	fmt.Println("Ping wait ", addr)
	var result ReplyInfo
	result.Cmd = cmd
	select {
	case rmsg := <-r:
		result.ReplyPkt = rmsg
		// else report protocol error
		fmt.Printf("ping read: %#v\n", rmsg)
		//time.Sleep(1 * time.Second)
	case <-time.After(1 * time.Second):
		// report timeout error
		fmt.Printf("Ping read timeout\n")
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
	fmt.Println("Looking for new devices..")
}

func FindDevices(r chan Message_t, w chan Message_t, ack_new_devices chan []uint32) {
	SendFindDevices(w)
	var new_devices []uint32
	timeout := time.After(1 * time.Second)
Loop:
	for {
		select {
		case <-timeout:
			fmt.Println("Find Devices Timeout")
			break Loop
		case rmsg := <-r:
			fmt.Println("Find Pkt Recv")
			if rmsg.Header.Mtype == NEED_ADDR { // and payload == "NEED ADDR"
				fmt.Println("Found A Device!")
				new_devices = append(new_devices, rmsg.Header.Mac)
			}
		}
	}
	ack_new_devices <- new_devices
	fmt.Println("Finish Find Devices")
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
		fmt.Println("Address assignment response not received.")
	}
	ack_reply <- msg
}

func (d GreenBus) runmaster(port string) {
	var (
		err          error
		orig_termios *termioslib.Termios
		ser          *os.File
	)

	defer func() {
		fmt.Println("Serial Exiting...")
		if err != nil {
			fmt.Println(err)
		}
	}()

	ser, err = os.OpenFile(port, os.O_RDWR, 777)
	if err != nil {
		return
	}

	defer func() {
		fmt.Println("Closing serial port")
		ser.Close()
	}()

	orig_termios, err = setup_serial(ser)

	if err != nil {
		return
	}

	defer func() {
		// set the settings back to the original when the program exits
		fmt.Println("Resetting termios")
		err = termioslib.Setattr(ser.Fd(), termioslib.TCSANOW, orig_termios)
	}()

	r := make(chan Message_t)
	w := make(chan Message_t)

	defer func() {
		close(r)
		close(w)
	}()

	defer func() {
		fmt.Println("STOPING DRIVER!!!!")
	}()

	go WriteMessages(ser, w)
	go ReadMessages(ser, r)
	fmt.Println("Resetting Client Addrs")
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
	for {
		// are we pinging devices?
		if state_ping_devices {
			fmt.Println("State Ping")
			// are we waiting for a reply? - no
			if ping_reply == nil {
				fmt.Println("State Ping Reply")
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
					fmt.Println("Cmds: ", devices[next_device+2].Cmds)
					var cmd Cmd_t
					if len(devices[next_device+2].Cmds) > 0 {
						cmd = devices[next_device+2].Cmds[0]
						devices[next_device+2].Cmds = devices[next_device+2].Cmds[1:]
					} else {
						cmd = Cmd_t{Payload: []byte("PING"), Mtype: PING, Rtype: PING_REPLY}
					}

					go PingDevice(r, w, ping_reply, devices[next_device+2].Addr, devices[next_device+2].Mac, cmd)
					fmt.Println("Cmds2: ", devices[next_device+2].Cmds)
					next_device += 1
				}
			}
		}
		if state_find_devices {
			fmt.Println("State Find")
			//are we waiting for new devices? - no
			if ack_new_devices == nil {
				// are we waiting for an ack reply? - no
				if ack_reply == nil {
					fmt.Println("State Find Ack")
					fmt.Println("New Devices ", new_devices)
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
							fmt.Println("Cmds: ", devices[next_device+2].Cmds)
							var cmd Cmd_t
							if len(devices[next_device+2].Cmds) > 0 {
								cmd = devices[next_device+2].Cmds[0]
								devices[next_device+2].Cmds = devices[next_device+2].Cmds[1:]
							} else {
								cmd = Cmd_t{Payload: []byte("PING"), Mtype: PING, Rtype: PING_REPLY}
							}

							go PingDevice(r, w, ping_reply, devices[next_device+2].Addr, devices[next_device+2].Mac, cmd)
							fmt.Println("Cmds2: ", devices[next_device+2].Cmds)
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
		/*case cmd := <-d.SendCmd:
		fmt.Println("Do Add To Queue")
		device := devices.Find(uint32(cmd.Device))
		device.Cmds = append(device.Cmds, cmd)*/
		case reply := <-ping_reply:
			fmt.Println("Do Ping Reply")
			ping_reply = nil
			if reply.ReplyPkt.Header.Mtype == reply.Cmd.Rtype {
				if reply.Cmd.ReplyHandler != nil {
					reply.Cmd.ReplyHandler(reply.ReplyPkt)
				}
			}
		case msg_devices := <-ack_new_devices:
			fmt.Println("Do Ack Devices")
			ack_new_devices = nil
			new_devices = msg_devices
		case rmsg := <-ack_reply:
			fmt.Println("Do Ack Reply")
			ack_reply = nil
			if rmsg.ReplyPkt.Header.Mtype == ACK_REPLY && rmsg.ReplyPkt.Header.Mac == rmsg.device_mac { // and payload == "ACK"
				interrogate_cmd := []Cmd_t{Cmd_t{Payload: []byte("INTERROGATE"),
					Mtype: INTERROGATE,
					Rtype: INTERROGATE_REPLY,
					ReplyHandler: func(msg Message_t) {
						fmt.Println("INTERROGATE REPLY")
						dev_properties := interrogate_payload(msg.Payload).ToStruct()
						devices[rmsg.device_addr].ModelName = dev_properties.Model
						devices[rmsg.device_addr].Serial = dev_properties.Serial

						fmt.Println("model: ", dev_properties.Model)
						fmt.Println("Serial: ", dev_properties.Serial)
						// parse the msg and init the ports
						//devices[new_device_addr].Ports
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
		fmt.Println("DEVICES: ", devices)
	}
}

func ReadMessages(c *os.File, r chan Message_t) {
	var buffer []byte
	var hsize int = binary.Size(Header_t{})
	in := make([]byte, 1)
	for {
		num_read, err := c.Read(in)
		if err != nil {
			fmt.Println(err)
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
					fmt.Println(err)
				}

				if checksum != header.Crc {
					fmt.Println("Header CRC doesn't match")
					buffer = buffer[1:]
					continue
				}
			}
			if len(buffer) > hsize+2 {
				iobuffer := new(bytes.Buffer)
				iobuffer.Write(buffer[1 : hsize+1])
				err = binary.Read(iobuffer, binary.LittleEndian, &header)
				if err != nil {
					fmt.Println(err)
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
		fmt.Printf("Msg To: %d Type: %d Length: %d\n", msg.Header.Destination, msg.Header.Mtype, msg.Header.Length)
		_, err := c.Write(ToByte(msg))
		if err != nil {
			fmt.Println(err)
			//return
		}
	}
}

func init() {
	fmt.Println("blah")
	network_manager.RegisterDriver("greenbus", GreenBus{StopChan: make(chan bool)})
	//go runmaster()
}

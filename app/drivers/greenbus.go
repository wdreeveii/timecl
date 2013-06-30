package drivers

import (
	"os"
	"time"
	"fmt"
	"bytes"
	"encoding/binary"
	//"github.com/robfig/revel"
	//"github.com/mewkiz/pkg/hashutil/crc16"
	//"github.com/robfig/revel/modules/jobs/app/jobs"
	"github.com/wdreeveii/termioslib"
	"timecl/app/network_manager"
)

type GreenBus struct {
    StopChan	chan bool
    GetChan	chan network_manager.GetDrvCmd
    SetChan	chan network_manager.SetDrvCmd
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
    return b
}

func (d GreenBus) Get(cmd network_manager.GetDrvCmd) {
    d.GetChan <- cmd
}

func (d GreenBus) Set(cmd network_manager.SetDrvCmd) {
    d.SetChan <- cmd
}

type Cmd_t struct {
    Payload	[]byte
    Mtype	uint8
    Rtype	uint8
    ReplyHandler func(Message_t)
    CmdInfo	interface{}
}

type PortFunction int
const (
    Input	PortFunction = iota
    Output
)

type Port_t struct {
    Type	PortFunction
    Value	interface{}
}

type GreenBusDevice struct {
    Addr	uint8
    Mac		uint32
    // things like number of errors
    Cmds	[]Cmd_t
    // device model
    Ports	[]Port_t
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

const (
    FIND_DEVICES	uint8 = 72
    NEED_ADDR		uint8 = 73
    ACK_DEVICE		uint8 = 74
    ACK_REPLY		uint8 = 75
    
    INTERROGATE		uint8 = 80
    INTERROGATE_REPLY	uint8 = 81
    
    PING 		uint8 = 44
    PING_REPLY 		uint8 = 45
    
    RESET_NETWORK 	uint8 = 58
    
    GET 		uint8 = 36
    GET_REPLY 		uint8 = 37
    SET 		uint8 = 38
    SET_REPLY 		uint8 = 39 
)

type Header_t struct {
    Destination uint8
    Mtype uint8
    Length uint16
    Mac uint32
    Crc uint16
}

type Message_t struct {
    Header Header_t
    Payload []byte
}

func crc16_update(crc uint16, a byte) uint16 {
    var i int
    crc ^= uint16(a)
    for i = 0; i < 8; i++ {
	if (crc & 1 != 0) {
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
    for i := 0; i < len(bytes) - 2; i++ {
	checksum = crc16_update(checksum, bytes[i])
    }
    return checksum
}

func (d GreenBus) runmaster(port string) {
    var (
	err error
	orig_termios termioslib.Termios
	work_termios termioslib.Termios
	ser *os.File
    )
    
    defer func () {
	fmt.Println("Serial Exiting...")
	if err != nil { 
	    fmt.Println(err)
	}
    }()
    
    ser, err = os.OpenFile(port, os.O_RDWR, 777)
    if err != nil { return }
    
    defer func () {
	fmt.Println("Closing serial port")
	ser.Close()
    }()
    
    // flush all buffers
    if err = termioslib.Flush(ser.Fd(), termioslib.TCIOFLUSH); err != nil { return }
    
    // save a copy of the original terminal configuration
    if err = termioslib.Getattr(ser.Fd(), &orig_termios); err != nil { return }

    // get a working copy
    if err = termioslib.Getattr(ser.Fd(), &work_termios); err != nil { return }

    work_termios.C_iflag &= ^(termioslib.BRKINT | termioslib.INPCK | termioslib.ISTRIP | termioslib.ICRNL | termioslib.IXON)
    work_termios.C_oflag &= ^(termioslib.OPOST | termioslib.ONLCR )
    work_termios.C_lflag &= ^(termioslib.ISIG | termioslib.ICANON | termioslib.IEXTEN | termioslib.ECHO | termioslib.ECHOE | termioslib.ECHOK )
    //work_termios.C_cflag &= ^(termioslib.CSIZE | termioslib.PARENB | termioslib.PARODD | termioslib.HUPCL | termioslib.CSTOPB )
    work_termios.C_cflag |= (termioslib.CS8)
    work_termios.C_cc[termioslib.VMIN] = 1
    work_termios.C_cc[termioslib.VTIME] = 0

    if err = termioslib.Setispeed(&work_termios, termioslib.B57600); err != nil { return }
    if err = termioslib.Setospeed(&work_termios, termioslib.B57600); err != nil { return }
    
    // set the working copy
    if err = termioslib.Setattr(ser.Fd(), termioslib.TCSANOW , &work_termios); err != nil { return }
    
    defer func () {
	// set the settings back to the original when the program exits
	fmt.Println("Resetting termios")
        err = termioslib.Setattr(ser.Fd(), termioslib.TCSANOW, &orig_termios)
    } ()
    
    r:= make(chan Message_t)
    w:= make(chan Message_t)
    
    defer func () {
	close(r)
	close(w)
    } ()
    
    defer func () {
	fmt.Println("STOPING DRIVER!!!!")
    } ()
    
    go WriteMessages(ser, w)
    go ReadMessages(ser, r)
    w <- Message_t{Payload: []byte("RESET ADDR"), Header: Header_t{Destination: 0, Mtype: RESET_NETWORK, Mac: 0}}

    var devices DeviceList = []GreenBusDevice{ GreenBusDevice{Addr: 0,Mac: 0},GreenBusDevice{Addr: 1,Mac: 0} }
    for {
SAVE_COMMANDS:
	for {
	    select {
	    case cmd := <- d.GetChan:
		for _, device := range devices {
		    if device.Mac == uint32(cmd.Device) {
			device.Cmds = append(device.Cmds, Cmd_t{Payload: []byte("GET ..."), Mtype: GET, Rtype: GET_REPLY, CmdInfo: cmd})
		    }
		}
	    case cmd := <- d.SetChan:
		for _, device := range devices {
		    if device.Mac == uint32(cmd.Device) {
			device.Cmds = append(device.Cmds, Cmd_t{Payload: []byte("SET ..."), Mtype: SET, Rtype: SET_REPLY, CmdInfo: cmd})
		    }
		}
	    default:
		break SAVE_COMMANDS
	    }
	}
//PING_DEVICES:
	for idx, device := range devices[2:] {
	    var cmd Cmd_t
	    if len(device.Cmds) > 0 {
		cmd = devices[2+idx].Cmds[0]
		devices[2+idx].Cmds = devices[2+idx].Cmds[1:]
	    } else {
		cmd = Cmd_t{Payload: []byte("PING"), Mtype: PING, Rtype: PING_REPLY}
	    }
	    
	    var msg Message_t
	    msg.Header.Destination = device.Addr
	    msg.Header.Mac = device.Mac
	    msg.Payload = cmd.Payload
	    msg.Header.Mtype = cmd.Mtype
	    
	    w <- msg
	    fmt.Println("beat")
	    select {
	    case <- d.StopChan:
		return
	    case rmsg := <- r:
		if rmsg.Header.Mtype == cmd.Rtype {
		    if cmd.ReplyHandler != nil {
			cmd.ReplyHandler(rmsg)
		    }
		}
		// else report protocol error
		fmt.Printf("read: %#v\n", rmsg)
		time.Sleep(1*time.Second)
	    case <- time.After(1*time.Second):
		// report timeout error
		fmt.Printf("read timeout\n")
		continue
	    }
	}
	var msg Message_t
	msg.Payload = []byte("HAHAHA")
	
	msg.Header.Destination = 0
	msg.Header.Mtype = FIND_DEVICES
	msg.Header.Mac = 0
	
	w <- msg
	fmt.Println("Looking for new devices..")
	var new_devices []uint32
FIND_DEVICES:
	for {
	    select {
	    case <- d.StopChan:
		return
	    case <- time.After(1*time.Second):
		break FIND_DEVICES
	    case rmsg := <- r:
		if rmsg.Header.Mtype == NEED_ADDR { // and payload == "NEED ADDR"
		    new_devices = append(new_devices, rmsg.Header.Mac)
		}
	    }
	}
//ACK_NEW_DEVICE:
	for _, val := range new_devices {
	    var msg Message_t
	    var device_addr uint8
	    
	    device_idx, device_in_list := devices.Find(val)
	    if !device_in_list {
		device_addr = uint8(len(devices))
	    } else {
		device_addr = uint8(device_idx)
	    }
	    
	    msg.Payload = []byte(fmt.Sprintf("%d", device_addr))
	    msg.Header.Destination = 0
	    msg.Header.Mtype = ACK_DEVICE
	    msg.Header.Mac = val

	    w <- msg
	    select {
	    case <- d.StopChan:
		return
	    case <- time.After(40*time.Millisecond):
		fmt.Println("Address assignment response not received.")
		continue
	    case rmsg := <- r:
		if rmsg.Header.Mtype == ACK_REPLY && rmsg.Header.Mac == val { // and payload == "ACK"
		    interrogate_cmd := []Cmd_t{Cmd_t{Payload: []byte("INTERROGATE"), 
					    Mtype: INTERROGATE, 
					    Rtype: INTERROGATE_REPLY,
					    ReplyHandler: func (msg Message_t) {
						fmt.Println("INTERROGATE REPLY")
						// parse the msg and init the ports
						//devices[new_device_addr].Ports
					    }}}
		    if !device_in_list {
			new_device := GreenBusDevice{Addr: device_addr,
							Mac: val,
							Cmds: interrogate_cmd}
			devices = append(devices, new_device)
		    } else {
			devices[device_addr].Cmds = append(interrogate_cmd, devices[device_addr].Cmds...)
		    }
		}
	    }
	}
	fmt.Println("DEVICES: ", devices)
    }
}

func ReadMessages(c *os.File, r chan Message_t) {
    var buffer []byte
    var hsize int = binary.Size(Header_t {})
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
	    if len(buffer) == hsize + 2 {
		if buffer[0] != 'A' || buffer[hsize + 1] != 'A' {
		    buffer = buffer[1:]
		    continue
		}
		var checksum uint16 = 0xffff
		for i:= 0; i < hsize - 1; i++ {
		    checksum = crc16_update(checksum, buffer[i])
		}
		iobuffer := new(bytes.Buffer)
		iobuffer.Write(buffer[1:hsize + 1])
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
	    if len(buffer) > hsize + 2 {
		iobuffer := new(bytes.Buffer)
		iobuffer.Write(buffer[1:hsize+1])
		err = binary.Read(iobuffer, binary.LittleEndian, &header)
		if err != nil {
		    fmt.Println(err)
		}
		if len(buffer) == hsize + 2 + int(header.Length) {
		    var checksum uint16 = 0xffff
		    for i := 0; i < len(buffer) - 2; i++ {
			checksum = crc16_update(checksum, buffer[i])
		    }
		    if checksum == binary.LittleEndian.Uint16(buffer[len(buffer) - 2:]) {
			var msg Message_t
			msg.Header = header
			msg.Payload = buffer[hsize + 2:len(buffer) - 2]
			r <- msg
		    }
		    buffer = make([]byte, 0)
		}
		if len(buffer) > hsize + 2 + int(header.Length) {
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
	msg, ok = <- w
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

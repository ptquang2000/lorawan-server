class FrameTable extends React.Component {
    state = {
        data: [],
        showOptions: [],
    }

    componentDidMount() {
        axios.get(this.props.path).then(res => {
            this.setState({ data: res.data })
        })
        this.props.setOptions(this.state.showOptions)
    }

    render() {
        return null
    }
}

const ByteBlocks = ({ size, value }) => {
    const base64ToHex = (str) => {
        for (var i = 0, bin = atob(str.replace(/[ \r\n]+$/, "")), hex = []; i < bin.length; ++i) {
            let tmp = bin.charCodeAt(i).toString(16)
            if (tmp.length === 1) tmp = "0" + tmp
            hex[hex.length] = tmp
        }
        return hex.join(" ")
    }
    const numberToHex = (number) => {
        for (var i = 0, hex = []; i < size; i++) {
            let tmp = (number & BigInt(0xFF)).toString(16).toUpperCase()
            tmp = tmp.length == 1 ? "0" + tmp : tmp
            hex[hex.length] = tmp 
            number = number >> BigInt(8)
        }
        return hex.join(" ")
    }

    let hex = 
        typeof(value) === 'string' ? base64ToHex(value) :
        typeof(value) === 'number' ? numberToHex(BigInt(value)) : null;

    return (
        <span 
            class="badge" 
            style={{
                color:"var(--bs-body-bg)", 
                backgroundColor:"var(--bs-light)"}}
        >{hex}</span>
    )
}

const EndDeviceTable = ({ path, showOptions, setOptions }) => {
    const [defaultOptions] = React.useState(["Appkey", "DevEui", "DevAddr"])
    const [data, setData] = React.useState([])

    React.useEffect(() => {
        axios.get(path).then(res => {
            setData(res.data)
        })
        setOptions(defaultOptions)
    }, [path])

    const formatBusId = (devEui) => {
        return devEui & 0xFFFFFF
    }

    const endevices  = data.map((device, index) => { return (
        <tr>
            <th scope="row">{index + 1}</th>
            <td>{formatBusId(device.Id)}</td>
            {showOptions['Appkey'] ? <td><ByteBlocks value={device.Appkey}/></td> : null}
            {showOptions['DevEui'] ? <td><ByteBlocks value={device.DevEui} size={8}/></td> : null}
            {showOptions['DevAddr'] ? <td><ByteBlocks value={device.DevAddr} size={4}/></td> : null}
            <td><button className="btn btn-link p-1"><ThreeVerticalDots/></button></td>
            <td><button className="btn btn-link p-1"><TrashFill/></button></td>
        </tr>
    )})

    return (
        <table class="table">
            <thead>
                <tr>
                <th scope="col">#</th>
                <th scope="col">Bus ID</th>
                {showOptions['Appkey'] ? <th scope="col">AppKey</th> : null}
                {showOptions['DevEui'] ? <th scope="col">DevEUI</th> : null}
                {showOptions['DevAddr'] ? <th scope="col">DevAddr</th> : null}
                <th scope="col">Actions</th>
                <th scope="col"></th>
                </tr>
            </thead>
            <tbody>{ endevices }</tbody>
        </table>
    )
}

const GatewayTable = ({ path, showOptions, setOptions }) => {

    const [defaultOptions] = React.useState(["Admin", 'Created Date'])
    const [data, setData] = React.useState([])

    React.useEffect(() => {
        axios.get(path).then(res => {
            setData(res.data)
        })
        setOptions(defaultOptions)
    }, [path])

    const gateways  = data.map((gateway, index) => { return (
        <tr>
            <th scope="row">{index + 1}</th>
            <td>{gateway.Username}</td>
            {showOptions['Admin'] ? 
            <td style={{color: gateway.Is_superuser ? "var(--bs-success)" : "var(--bs-danger)"}}>
                {gateway.Is_superuser ? <CheckIcon/> : <CrossIcon/>}
            </td> 
            : null}
            {showOptions['Created Date'] ? <td>{gateway.Created}</td> : null}
            <td><button className="btn btn-link p-1"><ThreeVerticalDots/></button></td>
            <td><button className="btn btn-link p-1"><TrashFill/></button></td>
        </tr>
    )})

    return (
        <table class="table">
            <thead>
                <tr>
                <th scope="col">#</th>
                <th scope="col">Station</th>
                {showOptions['Admin'] ? <th scope="col">Admin</th> : null}
                {showOptions['Created Date'] ? <th scope="col">Created Date</th> : null}
                <th scope="col">Actions</th>
                <th scope="col"></th>
                </tr>
            </thead>
            <tbody>{ gateways }</tbody>
        </table>
    )
}
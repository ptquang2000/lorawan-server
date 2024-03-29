const DataType = {
    Gateway : 'Gateway',
    EndDevices: 'EndDevices',
    Frames: 'Frames',
}

const Navbars = ( { onTabChanged } ) => {
    const tabs = [
        { type: DataType.Gateway, path: "gateways", text: "Gateways", isActive: true, icon: <RouterIcon/> },
        { type: DataType.EndDevices, path: "end-devices", text: "End Devices", isActive: false, icon: <BusIcon/> },
        { type: DataType.Frames, path: "frames/10", text: "Frames", isActive: false, icon: <ChatLeftText/> },
    ]

    const navbarHandle = (e) => {
        e.preventDefault()
        onTabChanged(e.target.href, e.target.ariaLabel)
    }

    const ref = React.useRef(null)
    const tabList = tabs.map(tab => {
        return (
            <li className="nav-item d-flex align-items-center">
                {tab.icon}
                <a
                ref={tab.isActive ? ref : null}
                aria-label={tab.type}

                className={"nav-link" + (tab.isActive ? " active" : "")} 
                {...(tab.isActive ? {"aria-current": "page"} : {})} 
                href={tab.path}

                onClick={navbarHandle}
                >{tab.text}
                </a>
            </li>
        )
    })

    React.useEffect(() => {
        onTabChanged(ref.current.href, ref.current.ariaLabel)
      }, [])

    return (
        <nav className="navbar align-items-start text-nowrap ps-3 pt-3" style={{backgroundColor:"var(--bs-info-bg-subtle)"}}>  
            <ul class="nav flex-column">
                { tabList }
            </ul>
        </nav>
    )
}

window.hu60_res_bot_icon = '/q.php/api.webplug-file/22780_public_hu_icon.svg'
window.hu60_res_back_icon = '/q.php/api.webplug-file/22780_public_return_icon.svg'
window.hu60_res_loading_icon = '/q.php/api.webplug-file/22780_public_loading.svg'
window.hu60_res_exit_chat_icon = '/q.php/api.webplug-file/22780_public_exit_window.svg'
window.hu60_res_clear_icon = '/q.php/api.webplug-file/22780_public_cancel_icon.svg'
window.hu60_res_clear_all_icon = '/q.php/api.webplug-file/22780_public_qingkong_icon.svg'
window.hu60_res_source_link_icon = '/q.php/api.webplug-file/22780_public_source_link_icon.svg'
window.hu60_res_default_avatar = '/upload/default.jpg'
window.hu60_res_robot_icon = 'https://file.hu60.cn/avatar/-50.jpg'
window.hu60_site_file_url = 'https://file.hu60.cn'
window.hu60_hu60bot_uid = -54
window.hu60_hu60bot_welcome = '您好，有什么我可以为您效劳的吗？'

function startPlugin() {

    initCurrentUserInfo()

    initHu60botChat()

    initGlobalListener()

    connectWs()

    showPluginDoor()

    smallScreenDeviceSafeInit()

    largeScreenDeviceSafeInit()
    
    checkUnreadMsgs()
}

function checkUnreadMsgs() {
    let formData = new FormData()
    formData.append('data', '{"type":1}')
    fetch('/q.php/api.msg.isread.get.json', {
        body: formData,
        method: "post"
    }).then(res => res.json()).then(jres => {
        console.log(JSON.stringify(jres))
        if(Object.keys(jres.result).length == 0) {
            document.querySelectorAll('#chatList li').forEach(chatItem => {
                let newMsgTips = chatItem.querySelector('.newMsgTips')
                newMsgTips.innerText = '0'
                newMsgTips.style.display = 'none'
            })
            let hu60botChatList = JSON.parse(window.localStorage.getItem("hu60bot_chat_list.json"))
            for(let i =0;i<hu60botChatList.length;i++) {
                hu60botChatList[i].tipsCount = 0
            }
            window.localStorage.setItem("hu60bot_chat_list.json", JSON.stringify(hu60botChatList))
        }
    })
}

function initCurrentUserInfo() {
    let uid = window.localStorage.getItem('uid')
    if(uid == null) {
        fetch('/q.php/user.index.json').then(r => r.json()).then(j => {
            window.localStorage.setItem('uid', j.uid)
            window.hu60_uid = j.uid
        })
    } else {
        window.hu60_uid = uid
    }
}

function initHu60botChat() {
    let hu60botChatBaseHTML = 
        `<div id="chatList">
            <div class="hltitle">
                <span class="hu60botwsstatus" title="disconnected"></span>
                <img src="${window.hu60_res_exit_chat_icon}" class="hu60botminwindow" title="minimize"/>
                <img src="${window.hu60_res_clear_all_icon}" class="hu60botclearall hu60botmenuicon" title="clear all conversation" />
            </div>
            <ul></ul>
        </div>
        <div id="chatWindow"></div>`
    
    let hu60botChatRoot = document.createElement('div')
    hu60botChatRoot.id = 'hu60botChat'
    hu60botChatRoot.innerHTML = hu60botChatBaseHTML

    document.body.appendChild(hu60botChatRoot)
  	
    document.querySelector('#hu60botChat .hu60botminwindow')
        .addEventListener('click', e => hu60botWindowOp(false))

    document.querySelector('#hu60botChat .hu60botclearall')
        .addEventListener('click', e => {
            JSON.parse(window.localStorage.getItem('hu60bot_chat_list.json')).forEach(chat => window.localStorage.removeItem(`${chat.uid}convo.json`))
            window.localStorage.removeItem('hu60bot_chat_list.json')
            initChatWindowData()
        })
    
    initChatWindowData()
}

function initChatWindowData() {
    renderChatList()
    document.querySelector(`#chat-${window.hu60_hu60bot_uid}`)
        .click()
}

function renderChatList() {
    document.querySelector("#chatList ul").innerHTML = ''
    let hu60bot_chat_list = window.localStorage.getItem("hu60bot_chat_list.json")
    if (hu60bot_chat_list == null) {
        let hu60botChat = {
            uid: window.hu60_hu60bot_uid,
            name: "hu60bot", 
            avatar: window.hu60_res_robot_icon, 
            isRobot: true,
            tipsCount: 0
        }
        window.localStorage.setItem("hu60bot_chat_list.json", JSON.stringify([hu60botChat]))
        appendChatList(hu60botChat, {updateStorage: false})
    } else {
        let hu60botChatList = JSON.parse(hu60bot_chat_list)
        hu60botChatList.forEach(chat => {
            if(chat.uid == window.hu60_hu60bot_uid) {
                appendChatList(chat, {updateStorage: false,focused: window.hu60_chatwindow == chat.uid})
            }
        })
        hu60botChatList.forEach(chat =>  {
            if(chat.uid == window.hu60_hu60bot_uid) {
                return
            }
            appendChatList(chat, {updateStorage: false,focused: window.hu60_chatwindow == chat.uid})
        })
    }
    if(window.innerWidth < 1080) {
        document.querySelectorAll('#chatList li')
            .forEach(chatItem => chatItem.addEventListener('click', e=>{
                document.querySelector('#chatWindow').style.display = 'block'
                document.querySelector('#chatList').style.display = 'none'
            }))
    }
}

function appendChatList(chat, opts={updateStorage: true, focused: false}) {
    // UI
    const chatItem = document.createElement('li')
    chatItem.innerHTML = 
        `<span class="newMsgTips">${chat.tipsCount}</span>
        <img class="cavatar" src="${chat.avatar}" />
        ${chat.name}<br />
        <span class="latestMsgOverview">${chat.overview?chat.overview:""}</span>
        <img class="clearChat" src="${window.hu60_res_clear_icon}" />`
    chatItem.id = `chat-${chat.uid}`
    chatItem.addEventListener('mouseover', e => chatItem.querySelector('.clearChat').style.display = 'block')
    chatItem.addEventListener('mouseout', e => chatItem.querySelector('.clearChat').style.display = 'none')
    chatItem.addEventListener('click', e => {
        document.querySelectorAll('#chatList li')
            .forEach( item => item.classList.remove('activeChat'))
        chatItem.classList.add("activeChat")
        initChatWindow(chat)
        let newMsgTips = chatItem.querySelector('.newMsgTips')
        newMsgTips.innerText = '0'
        newMsgTips.style.cssText = 'display: none'
        let hu60botChatList = JSON.parse(window.localStorage.getItem("hu60bot_chat_list.json"))
        for(let i =0;i<hu60botChatList.length;i++) {
            let chat = hu60botChatList[i]
            if(chat.uid == window.hu60_chatwindow) {
                chat.tipsCount = 0
                break
            }
        }
        window.localStorage.setItem("hu60bot_chat_list.json", JSON.stringify(hu60botChatList))
    })

    chatItem.querySelector('.clearChat').addEventListener('click', e=> {
        if(chat.isRobot) {
            window.localStorage.removeItem(`${chat.uid}convo.json`)
            document.querySelector('#chatContainer').innerHTML = ''
            appendChatText(window.hu60_hu60bot_welcome, chat.uid, 
                {self: false, updateStorage: true})

            if(chat.uid == window.hu60_hu60bot_uid) {
                window.hu60_ws.send(JSON.stringify({action: "rmconvo"}))
            }
            e.stopPropagation()
            return
        }
        let prevUid = window.hu60_chatwindow
        let hu60botChatList = JSON.parse(window.localStorage.getItem('hu60bot_chat_list.json'))
        for(let i =0;i<hu60botChatList.length;i++) {
            if(hu60botChatList[i].uid == chat.uid) {
                let elementToMove = hu60botChatList.splice(i, 1)
                window.localStorage.setItem('hu60bot_chat_list.json', JSON.stringify(hu60botChatList))
                break
            }
        }
        renderChatList()
        e.stopPropagation()
    })

    if (opts.focused) {
        document.querySelectorAll('#chatList li')
            .forEach( item => item.classList.remove('activeChat'))
        chatItem.classList.add('activeChat')
    }

    if (chat.tipsCount > 0) {
        chatItem.querySelector('.newMsgTips').style.cssText = 'display: block'
    }

    document.querySelector("#chatList ul").appendChild(chatItem)

    // Update data
    if(opts.updateStorage) {
        let hu60bot_chat_list = window.localStorage.getItem("hu60bot_chat_list.json")
        let hu60botChatList = null
        if (hu60bot_chat_list == null) {
            hu60botChatList = []
        } else {
            hu60botChatList = JSON.parse(hu60bot_chat_list)
        }
        if(hu60botChatList[0].uid == chat.uid) {
            return
        }
        hu60botChatList.unshift(chat)
        window.localStorage.setItem("hu60bot_chat_list.json", JSON.stringify(hu60botChatList))
    }
}

function initChatWindow(chat) {
    if(window.hu60_chatwindow) {
        // update UI
        window.hu60_chatwindow = chat.uid
        window.hu60_chatwindow_obj = chat
        document.querySelector("#ctitle .chatName").innerText = chat.name
        document.querySelector('#chatContainer').innerHTML = ''
        let convo = window.localStorage.getItem(`${chat.uid}convo.json`)
        if (convo == null) {
            if (chat.isRobot) {
                appendChatText(window.hu60_hu60bot_welcome, chat.uid, 
                    {self: false, updateStorage: true})
            }
        } else {
            JSON.parse(convo).forEach(c => {
                appendChatText(c.words, chat.uid, 
                    {self: c.self, updateStorage: false,msgid: c.msgid})  
            })
        }
        return
    }
    // create UI
    window.hu60_chatwindow = chat.uid
    window.hu60_chatwindow_obj = chat

    let chatWindow = document.querySelector("#chatWindow")


    const chatReturnBtn = document.createElement('img')
    chatReturnBtn.id = 'chatReturnBtn'
    chatReturnBtn.src = window.hu60_res_back_icon
    chatReturnBtn.addEventListener('click', e => {
        document.querySelector('#chatWindow').style.display = 'none'
        document.querySelector('#chatList').style.display = 'block'
    })
    chatWindow.appendChild(chatReturnBtn)


    const chatTitile = document.createElement("div")
    chatTitile.id = 'ctitle'
    chatTitile.innerHTML = `<span class="chatName">${chat.name}</span>`
    chatWindow.appendChild(chatTitile)


    const chatContainer = document.createElement("ul")
    chatContainer.id = 'chatContainer'
    chatWindow.appendChild(chatContainer)


    const chatInput = document.createElement("div")
    chatInput.id = 'chatInput'

    const _chatInputSend = document.createElement("button")
    _chatInputSend.innerText="发送"
    _chatInputSend.style.display = 'none'
    _chatInputSend.id = 'chatInputSend'
    _chatInputSend.addEventListener("click", e => {
        let text = document.querySelector("#chatInput textarea")
        text.focus()
        if (text.value == "") {
            text.style.cssText = "border: 1px solid red"
            setTimeout(()=>{text.style.cssText = "border: 0.5px solid #ddd"},200)
            setTimeout(()=>{text.style.cssText = "border: 1px solid red"},300)
            setTimeout(()=>{text.style.cssText = "border: 0.5px solid #ddd"},500)
            setTimeout(()=>{text.style.cssText = "border: 1px solid red"},600)
            setTimeout(()=>{text.style.cssText = "border: 0.5px solid #ddd"},800)
            return
        }
        handleMsgSend(text.value)
        appendChatText(text.value, window.hu60_chatwindow, {self: true, updateStorage: true})
        text.value = ''
        _chatInputSend.style.display = 'none'
    })

    const _chatInput = document.createElement("textarea")
    _chatInput.placeholder="请提问（Ctrl + Enter 快捷发送）"
    _chatInput.addEventListener("input", e=>{_chatInputSend.style.display = 'block'})

    chatInput.appendChild(_chatInput)
    chatInput.appendChild(_chatInputSend)
    
    chatWindow.appendChild(chatInput)

    // render conversation
    let convo = window.localStorage.getItem(`${chat.uid}convo.json`)
    if (convo == null) {
        if (chat.isRobot) {
            appendChatText(window.hu60_hu60bot_welcome, chat.uid, 
                {self: false, updateStorage: true})
        }
    } else {
        JSON.parse(convo).forEach(c => 
            appendChatText(c.words, chat.uid, 
                {self: c.self, updateStorage: false,msgid:c.msgid}))
    }
}

function appendChatText(words, uid, opts={self: false,updateStorage: true}) {
    if (opts.updateStorage) {
        // update convo.json and chat_list.json (for order the chat list)
        let convo = window.localStorage.getItem(`${uid}convo.json`)
        let convoObj = JSON.parse(convo)
        if (convo == null) {
            convoObj = []
        }

        let isRepeat = false
        if(convoObj.length > 0) {
            let lastConvo = convoObj.slice(-1)[0]
            if(lastConvo.words == words && lastConvo.self == opts.self) {
                isRepeat = true
            }
        }
        
        if(!isRepeat) {
            convoObj.push({words: words,self: opts.self, msgid: opts.msgid})
            window.localStorage.setItem(`${uid}convo.json`, JSON.stringify(convoObj))
        }
        
        document.querySelector(`#chat-${uid} .latestMsgOverview`).innerText = words
        if(window.hu60_chatwindow != uid) {
            let newMsgTips = document.querySelector(`#chat-${uid} .newMsgTips`)
            let newTipsCount = parseInt(newMsgTips.innerText)+1
            newMsgTips.innerText = newTipsCount
            if(newTipsCount>0) {
                newMsgTips.style.display = 'block'
            }
        }
        let hu60bot_chat_list = window.localStorage.getItem('hu60bot_chat_list.json')
        if (hu60bot_chat_list != null) {
            let hu60botChatList = JSON.parse(hu60bot_chat_list)
            let hit = null
            for(let i =0 ;i < hu60botChatList.length;i++) {
                let chat = hu60botChatList[i]
                if (chat.uid == uid ) {
                    hit = i
                    chat.overview = words
                    if(window.hu60_chatwindow != uid) {
                        chat.tipsCount = 
                            parseInt(document.querySelector(`#chat-${uid} .newMsgTips`).innerText)
                    }
                    hu60botChatList[i] = chat
                    break
                }
            }
            let elementToMove = hu60botChatList.splice(hit, 1)
            hu60botChatList.unshift(elementToMove[0])
            window.localStorage.setItem('hu60bot_chat_list.json', JSON.stringify(hu60botChatList))
        }
        renderChatList()
    }

    if(window.hu60_chatwindow != uid) {
        return
    }
    // create UI (text, avatar and loading icon)
    const chatItem = document.createElement("li")
    chatItem.classList.add('chat')

    let chatContainer = document.querySelector('#chatContainer')
    chatContainer.appendChild(chatItem)

    const avatar = document.createElement("img")
    avatar.classList.add('cavatar')
    avatar.src = window.hu60_chatwindow_obj.avatar

    const text = document.createElement("div")
    text.classList.add('hu60bot')
    let optComp = ""
    if(opts.msgid) {
        optComp = `<a class="source-link" href="/q.php/link.ack.msg.${opts.msgid}.html" title="click me">
            <img src="${window.hu60_res_source_link_icon}" /></a>`
    }
    text.innerHTML = `${words}${optComp}`

    chatItem.appendChild(avatar)
    chatItem.appendChild(text)

    if (opts.self) {
        avatar.classList.add('crightpos')
        text.classList.add('crightpos')
        text.classList.add('cuser')
        text.innerText = words
        avatar.src = `${window.hu60_site_file_url}/avatar/${window.hu60_uid}.jpg`
        avatar.addEventListener('error', (e) => avatar.src = window.hu60_res_default_avatar)

        if(opts.updateStorage) {
            const icon = document.createElement("img")
            icon.classList.add('send_status_icon')
            icon.classList.add('crightpos')
            icon.src = window.hu60_res_loading_icon
            chatItem.appendChild(icon)
        }
    } else {
        if(opts.updateStorage) {
            document.querySelectorAll('.send_status_icon')
                .forEach(icon => icon.style.display = 'none')
        }
    }
    chatContainer.scrollTop = chatContainer.scrollHeight
}

function hu60botWindowOp(open) {
    if(open) {
      	document.querySelector('header').style.display = 'none'
        document.querySelector('.container').style.display = 'none'
        document.querySelector('footer').style.display = 'none'
        document.querySelector('#hu60botChat').style.display = 'flex'
        let chatContainer = document.querySelector('#chatContainer') 
        chatContainer.scrollTop = chatContainer.scrollHeight
        return
    }
  	document.querySelector('header').style.display= 'block'
    document.querySelector('.container').style.display= 'block'
    document.querySelector('footer').style.display= 'block'
    document.querySelector('#hu60botChat').style.display = 'none'
}

function initGlobalListener() {
    document.addEventListener("keydown", function(event){
        if(event.ctrlKey && event.code === 'Enter') {
            document.querySelector('#chatInput button').click()
        }
    })

    setInterval(()=>{
        const tips = document.querySelectorAll('#chatList .newMsgTips')
        for(let i =0;i<tips.length;i++) {
            let tipsCount = parseInt(tips[i].innerText)
            if(tipsCount>0) {
                if(window.hu60_new_tips_task) {
                    return
                }
                window.hu60_new_tips_task = setInterval(()=>{
                    const hu60botPlugin = document.querySelector('#hu60botPlugin')
                    if(hu60botPlugin.style.opacity == 0) {
                        hu60botPlugin.style.opacity = 1
                    } else {
                        hu60botPlugin.style.opacity = 0
                    }
                },500)
                return
            }
        }
        clearInterval(window.hu60_new_tips_task)
        document.querySelector('#hu60botPlugin').style.opacity = 1
        window.hu60_new_tips_task = null
    },100)
}

function connectWs() {
    let protocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
    let ws = new WebSocket(`${protocol}://${location.host}/ws/msg`)
    window.hu60_ws = ws
    ws.addEventListener("message", (event) => handleWsMsg(JSON.parse(event.data)))
  	ws.addEventListener("error", e => window.hu60_ws.close())
    ws.addEventListener("open", e => {
        window._hu60bot_hb_task = setInterval(() => ws.send('{"action":"ping"}'), 60000)
        let wsStatus = document.querySelector('#chatList .hu60botwsstatus')
        wsStatus.style.cssText = 'background: green'
        wsStatus.title = 'connected'
    })
    ws.addEventListener("close", e => {
        clearInterval(window._hu60bot_hb_task)
        let wsStatus = document.querySelector('#chatList .hu60botwsstatus')
        wsStatus.style.cssText = 'background: red'
        wsStatus.title = 'WS_CLOSED'
        setTimeout(() => window.hu60_ws = connectWs(), 5000)
    })
}

function showPluginDoor() {
    let floatPluginMenu = document.querySelector('#floatPluginMenu')
    if (floatPluginMenu == null) {
        floatPluginMenu = document.createElement("div")
        floatPluginMenu.id = 'floatPluginMenu'
        document.querySelector('body').appendChild(floatPluginMenu)
    }
    let hu60botPlugin = document.createElement("div")
    hu60botPlugin.id = 'hu60botPlugin'
    hu60botPlugin.innerHTML = `<img src="${window.hu60_res_bot_icon}" />`
    hu60botPlugin.addEventListener('click', e => {
        let hu60botChat = document.querySelector('#hu60botChat')
        if (hu60botChat.style.display == 'none' || hu60botChat.style.display == "") {
            hu60botWindowOp(true)
        } else {
            hu60botWindowOp(false)
        }
    })
    floatPluginMenu.appendChild(hu60botPlugin)
}

async function handleMsgSend(words) {
    let currentChatWindowUID = window.hu60_chatwindow
    if (!currentChatWindowUID) {
        alert("SYS_ERR")
        return
    }
    if(currentChatWindowUID == window.hu60_hu60bot_uid) {
        try {
            window.hu60_ws.send(JSON.stringify({action: "chat", data: words, id: "chat"}))
        } catch(e) {
          	log.debug(JSON.stringify(e))
            appendChatText("WS_SEND_ERR: " + JSON.stringify(e), window.hu60_chatwindow, 
                {self: false,updateStorage: false})
            try{
                if (window.hu60_ws.readyState === WebSocket.CLOSED) {
                    setTimeout(() => 
                        window.hu60_ws.send(JSON.stringify({action: "chat", data: text.value})), 6000)
                }
            } catch(_){}
        }
        return
    }

    // 发送聊天室消息
    let chatroom = await (await fetch('/q.php/addin.chat.hu60bot.json')).json()
    let formData = new FormData()
    formData.append('content', '<!md>\n@#' + currentChatWindowUID + ' ' + words)
    formData.append('token', chatroom.token)
    formData.append('go', '1')
    fetch('/q.php/addin.chat.hu60bot.json', {
        body: formData,
        method: "post",
        redirect: "manual" // 不自动重定向
    }).then(res => res.json()).then(jres => {
        console.log(JSON.stringify(jres))
        if(jres.success) {
            document.querySelectorAll('.send_status_icon').forEach(icon => icon.style.display='none')
        }
    })
}

function handleWsMsg(msg) {
    if (msg.event == 'chat') {
        appendChatText(`${msg.data.newConversation?"[新会话] ":""}${setext(msg.data.response)}`,
            window.hu60_hu60bot_uid, 
            {self: false, updateStorage: true})
        return
    }

    if(msg.event == 'ack') {
        if(msg.data == 'chat') {
            document.querySelectorAll('.send_status_icon').forEach(icon => icon.style.animationDuration = '0.5s')
        }
        return
    }

    if (msg.event == 'msg' && msg.data.type == 1) {
        let msgContent = JSON.parse(msg.data.content)
        let msgContentText = getHu60MsgText(msgContent[0].msg)

        if (document.querySelector(`#chat-${msg.data.byuid}`) == null) {
            if (!window.hu60_user_info_map) {
                window.hu60_user_info_map = {}
            }
            if (!window.hu60_user_info_map[msg.data.byuid]) {
                fetch(`/q.php/user.info.${msg.data.byuid}.json`)
                .then(res => res.json()).then(jres => {
                    window.hu60_user_info_map[msg.data.byuid] = jres
                    appendChatList({
                        uid: msg.data.byuid,
                        name: jres.name, 
                        avatar: `${window.hu60_site_file_url}/avatar/${msg.data.byuid}.jpg`, 
                        isRobot: msg.data.byuid<0,
                        overview: msgContentText,
                        tipsCount: 0
                    }, {updateStorage: true})
                    appendChatText(msgContentText, msg.data.byuid, 
                        {self: false, updateStorage: true, msgid: msg.data.id})
                })
                return
            }
            let userInfo = window.hu60_user_info_map[msg.data.byuid]
            appendChatList({
                uid: msg.data.byuid,
                name: userInfo.name, 
                avatar: `${window.hu60_site_file_url}/avatar/${msg.data.byuid}.jpg`, 
                isRobot: msg.data.byuid<0,
                overview: msgContentText,
                tipsCount: 0
            }, {updateStorage: true})
        }
        appendChatText(msgContentText, msg.data.byuid, 
            {self: false, updateStorage: true, msgid: msg.data.id})
        return
    }

    console.debug('unsupported message: ', msg)
}

function smallScreenDeviceSafeInit() {
    if (window.innerWidth <= 1080) {
        let hu60botChat = document.querySelector('#hu60botChat')
        let chatList = document.querySelector('#chatList')
        let chatContainer = document.querySelector('#chatContainer')
        let chatReturnBtn = document.querySelector('#chatReturnBtn')

        let winHeight = window.innerHeight


        hu60botChat.style.cssText = `height: ${winHeight}px`
        chatContainer.style.cssText = `height: ${(winHeight - 130)}px`

        chatList.style.display = 'none'
        chatReturnBtn.style.display = 'block'

        document.querySelectorAll('#chatList li')
            .forEach(chatItem => chatItem.addEventListener('click', e=>{
                document.querySelector('#chatWindow').style.display = 'block'
                document.querySelector('#chatList').style.display = 'none'
            }))
    }
}

function largeScreenDeviceSafeInit() {
    if (window.innerWidth > 1080) {
        let hu60botChat = document.querySelector('#hu60botChat')
        if(window.innerHeight<=720) {
            hu60botChat.style.bottom = '0px'
        } else {
            hu60botChat.style.bottom = `${(window.innerHeight-720)/2}px`
        }
    }
}


startPlugin()

// -----
// utils
// -----

function getHu60MsgText(msgContent) {
    const validUnits = ["text","imgzh","mdpre","mdcode","face","at"]
    let text = msgContent.filter(unit => validUnits.includes(unit.type) ).map(unit => {
        if(unit.type == "text") {
            return setext(unit.value)
        }
        if(unit.type == "mdpre") {
            return unit.data
        }
        if(unit.type == "face") {
            return `{${unit.face}}`
        }
        if(unit.type == "mdcode") {
            return unit.quote + unit.lang + unit.data + unit.quote
        }
        if(unit.type == "at") {
            return `@${unit.tag}(${unit.uid}) `
        }
        if(unit.type == "imgzh") {
            return `<img style="max-width: 90%;display: block" src="${unit.src}" alt="${unit.alt}" />`
        }
    }).join("").trim()
    for(let i=0;i<3;i++) {
        if(text.startsWith('<br />')) {
            text = text.substr(6)
        }
    }
    for(let i =0;i<5;i++) {
        if (text.startsWith(',') || text.startsWith('，') || text.startsWith('\n')) {
            text = text.substr(1)
        }
    }
    console.debug(text)
  return text
}

function setext(unsestr) {
    return unsestr.replaceAll('<', '&lt;').replaceAll('>','&gt;').replaceAll(' ', '&nbsp;').replaceAll('\n', '<br />')
}

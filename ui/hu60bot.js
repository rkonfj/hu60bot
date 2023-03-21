window.hu60_res_bot_icon = '/q.php/api.webplug-file/22780_public_hu_icon.svg'
window.hu60_res_back_icon = '/q.php/api.webplug-file/22780_public_return_icon.svg'
window.hu60_res_loading_icon = '/q.php/api.webplug-file/22780_public_loading.svg'
window.hu60_res_exit_chat_icon = '/q.php/api.webplug-file/22780_public_exit_window.svg'
window.hu60_res_robot_icon = 'https://file.hu60.cn/avatar/-50.jpg'
window.hu60_site_url = 'https://hu60.cn'
window.hu60_site_file_url = 'https://file.hu60.cn'
window.hu60_hu60bot_uid = -54

function startPlugin() {

    initCurrentUserInfo()

    initHu60botChat()

    initGlobalListener()

    connectWs()

    showPluginDoor()

    if (window.innerWidth <= 1080) {
        onSmallScreenDevice()
    } 
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
    '<div id="hu60botChat"><div id="chatList"><div class="hltitle"><span class="hu60botwsstatus" title="disconnected"></span><img src="'
    +window.hu60_res_exit_chat_icon+'" class="hu60botminwindow" title="minimize"/></div><ul></ul></div><div id="chatWindow"></div></div>'
    document.body.innerHTML+=hu60botChatBaseHTML
  	document.querySelector('#hu60botChat .hu60botminwindow').addEventListener('click', e => hu60botWindowOp(false))
    renderChatList()
    document.querySelector('#chat--54').click()

    // chatWindow
    let chatWindow = document.querySelector("#chatWindow")

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
}

function appendChatList(chat, opts={updateStorage: true, focused: false}) {
    const chatItem = document.createElement('li')
    chatItem.innerHTML = '<span class="newMsgTips">'+chat.tipsCount+'</span><img class="cavatar" src="'+chat.avatar+'" />'
        +chat.name+'<br /><span class="latestMsgOverview">'
        +(chat.overview?chat.overview:"")+'</span>'
    chatItem.id = 'chat-' + chat.uid
    if (opts.focused) {
        document.querySelectorAll('#chatList li').forEach( item => item.classList.remove('activeChat'))
        chatItem.classList.add('activeChat')
    }
    chatItem.addEventListener('click', e => {
        document.querySelectorAll('#chatList li').forEach( item => item.classList.remove('activeChat'))
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

    if (chat.tipsCount > 0) {
        chatItem.querySelector('.newMsgTips').style.cssText = 'display: block'
    }

    let chatList = document.querySelector("#chatList ul")
    chatList.appendChild(chatItem)
    if(opts.updateStorage) {
        let hu60bot_chat_list = window.localStorage.getItem("hu60bot_chat_list.json")
        let hu60botChatList = null
        if (hu60bot_chat_list == null) {
            hu60botChatList = []
        } else {
            hu60botChatList = JSON.parse(hu60bot_chat_list)
        }
        hu60botChatList.unshift(chat)
        window.localStorage.setItem("hu60bot_chat_list.json", JSON.stringify(hu60botChatList))
    }
}

function initChatWindow(chat) {
    if(window.hu60_chatwindow) {
        window.hu60_chatwindow = chat.uid
        window.hu60_chatwindow_obj = chat
        document.querySelector("#ctitle .chatName").innerText = chat.name
        document.querySelector('#chatContainer').innerHTML = ''
        let convo = window.localStorage.getItem(chat.uid + "convo.json")
        if (convo == null) {
            if (chat.isRobot) {
                appendChatText('您好，有什么我可以为您效劳的吗？', chat.uid, {self: false, updateStorage: true})
            }
        } else {
            JSON.parse(convo).forEach(c => {
                appendChatText(c.words, chat.uid, {self: c.self, updateStorage: false})  
            })
        }
        return
    }
    window.hu60_chatwindow = chat.uid
    window.hu60_chatwindow_obj = chat
    const chatReturnBtn = document.createElement('img')
    chatReturnBtn.id = 'chatReturnBtn'
    chatReturnBtn.src = window.hu60_res_back_icon
    chatReturnBtn.addEventListener('click', e => {
        document.querySelector('#chatWindow').style.display = 'none'
        document.querySelector('#chatList').style.display = 'block'
    })

    const chatTitile = document.createElement("div")
    chatTitile.id = 'ctitle'
    chatTitile.innerHTML = '<span class="chatName">' + chat.name + '</span>'

    const chatContainer = document.createElement("ul")
    chatContainer.id = 'chatContainer'

    const chatInput = document.createElement("div")
    chatInput.id = 'chatInput'
    const _chatInputSend = document.createElement("button")
    _chatInputSend.innerText="发送"
    _chatInputSend.style.cssText = 'display: none'
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
        handleSendMsg(text.value)
        appendChatText(text.value, window.hu60_chatwindow, {self: true, updateStorage: true})
        text.value = ''
        _chatInputSend.style.cssText = 'display: none'
    })
    const _chatInput = document.createElement("textarea")
    _chatInput.placeholder="请提问（Ctrl + Enter 快捷发送）"
    _chatInput.addEventListener("input", e=>{_chatInputSend.style.cssText = 'display: block'})


    chatInput.appendChild(_chatInput)
    chatInput.appendChild(_chatInputSend)

    let chatWindow = document.querySelector("#chatWindow")
    chatWindow.appendChild(chatReturnBtn)
    chatWindow.appendChild(chatTitile)
    chatWindow.appendChild(chatContainer)
    chatWindow.appendChild(chatInput)


    let convo = window.localStorage.getItem(chat.uid + "convo.json")
    if (convo == null) {
        if (chat.isRobot) {
            appendChatText('您好，有什么我可以为您效劳的吗？', chat.uid, {self: false, updateStorage: true})
        }
    } else {
        JSON.parse(convo).forEach(c => {
            appendChatText(c.words, chat.uid, {self: c.self, updateStorage: false})  
        })
    }
}

function appendChatText(words, uid, opts={self: false,updateStorage: true}) {
    if (opts.updateStorage) {
        let convo = window.localStorage.getItem(uid + "convo.json")
        let convoObj = JSON.parse(convo)
        if (convo == null) {
            convoObj = []
        }
        convoObj.push({words: words,self: opts.self})
        window.localStorage.setItem(uid + "convo.json", JSON.stringify(convoObj))

        document.querySelector('#chat-' + uid + ' .latestMsgOverview').innerText = words
        if(window.hu60_chatwindow != uid) {
            let newMsgTips = document.querySelector('#chat-' + uid + ' .newMsgTips')
            let newTipsCount = parseInt(newMsgTips.innerText)+1
            newMsgTips.innerText = newTipsCount
            if(newTipsCount>0) {
                newMsgTips.style.cssText = 'display: block'
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
                        chat.tipsCount = parseInt(document.querySelector('#chat-' + uid + ' .newMsgTips').innerText)
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

    const chatItem = document.createElement("li")
    chatItem.classList.add('chat')

    let chatContainer = document.querySelector('#chatContainer')
    chatContainer.appendChild(chatItem)

    const avatar = document.createElement("img")
    avatar.classList.add('cavatar')
    avatar.src = window.hu60_chatwindow_obj.avatar

    const text = document.createElement("div")
    text.classList.add('hu60bot')
    text.innerHTML = words

    chatItem.appendChild(avatar)
    chatItem.appendChild(text)

    if (opts.self) {
        avatar.classList.add('crightpos')
        text.classList.add('crightpos')
        text.classList.add('cuser')
        avatar.src = 'https://file.hu60.cn/avatar/'+window.hu60_uid+'.jpg'

        if(opts.updateStorage) {
            const icon = document.createElement("img")
            icon.classList.add('send_status_icon')
            icon.classList.add('crightpos')
            icon.src = window.hu60_res_loading_icon
            chatItem.appendChild(icon)
        }
    } else {
        if(opts.updateStorage) {
            document.querySelectorAll('.send_status_icon').forEach(icon => icon.style.cssText='display: none')
        }
    }
    chatContainer.scrollTop = chatContainer.scrollHeight
}

function hu60botWindowOp(open) {
    if(open) {
      	document.querySelector('header').style.cssText= 'display: none'
        document.querySelector('.container').style.cssText= 'display: none'
        document.querySelector('footer').style.cssText= 'display: none'
        document.querySelector('#hu60botChat').style.display = 'flex'
        let chatContainer = document.querySelector('#chatContainer') 
        chatContainer.scrollTop = chatContainer.scrollHeight
        return
    }
  	document.querySelector('header').style.cssText= 'display: block'
    document.querySelector('.container').style.cssText= 'display: block'
    document.querySelector('footer').style.cssText= 'display: block'
    document.querySelector('#hu60botChat').style.display = 'none'
}


function initGlobalListener() {
    // keydown
    document.addEventListener("keydown", function(event){
        if(event.ctrlKey && event.code === 'Enter') {
            document.querySelector('#chatInput button').click()
        }
    })
}


function connectWs() {
    let protocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
    let ws = new WebSocket(protocol+"://"+location.host+"/ws/msg")
    window.hu60_ws = ws
    ws.addEventListener("message", (event) => {
        let msg = JSON.parse(event.data)
        handleWsMsg(msg)
    })
    ws.addEventListener("open", e => {
        window._hu60bot_hb_task = setInterval(() => ws.send('{"action":"ping"}'), 60000)
        let wsStatus = document.querySelector('#chatList .hu60botwsstatus')
        wsStatus.style.cssText = 'background: green'
        wsStatus.title = 'connected'
    })
  	ws.addEventListener("error", e => {
     	window.hu60_ws.close()
    })
    ws.addEventListener("close", e => {
        clearInterval(window._hu60bot_hb_task)
        let wsStatus = document.querySelector('#chatList .hu60botwsstatus')
        wsStatus.style.cssText = 'background: red'
        wsStatus.title = 'WS_CLOSED'
        setTimeout(() => {
            window.hu60_ws = connectWs()
        }, 5000)
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
    hu60botPlugin.innerHTML = '<img src="'+window.hu60_res_bot_icon+'" />'
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


async function handleSendMsg(words) {
    let currentChatWindowUID = window.hu60_chatwindow
    if (!currentChatWindowUID) {
        alert("SYS_ERR")
        return
    }
    if(currentChatWindowUID == window.hu60_hu60bot_uid) {
        try {
            window.hu60_ws.send(JSON.stringify({action: "chat", data: words}))
        } catch(e) {
          	log.debug(JSON.stringify(e))
            appendChatText("WS_SEND_ERR: " + JSON.stringify(e), window.hu60_chatwindow, {self: false,updateStorage: false})
            try{
                if (window.hu60_ws.readyState === WebSocket.CLOSED) {
                    setTimeout(()=>{window.hu60_ws.send(JSON.stringify({action: "chat", data: text.value}))}, 6000)
                }
            } catch(_){}
        }
        return
    }

    // 发送聊天室消息

    let chatroom = await (await fetch(window.hu60_site_url + '/q.php/addin.chat.hu60bot.json')).json()

    let formData = new FormData()
    formData.append('content', '<!md>\n@#' + currentChatWindowUID + ' ' + words)
    formData.append('token', chatroom.token)
    formData.append('go', '1')
    fetch(window.hu60_site_url + '/q.php/addin.chat.hu60bot.json', {
        body: formData,
        method: "post",
        redirect: "manual" // 不自动重定向
    }).then(res => res.json()).then(jres => {
        console.log(JSON.stringify(jres))
        if(jres.success) {
            document.querySelectorAll('.send_status_icon').forEach(icon => icon.style.cssText='display: none')
        }
    })
    
}


function handleWsMsg(msg) {
    if (msg.event == 'chat') {
        appendChatText((msg.data.newConversation?"[新会话]":"") + setext(msg.data.response), window.hu60_hu60bot_uid, {self: false, updateStorage: true})
        return
    }

    if (msg.event == 'msg' && msg.data.type == 1) {
        let msgContent = JSON.parse(msg.data.content)
        let msgContentText = getHu60MsgText(msgContent[0].msg)

        if (document.querySelector('#chat-'+msg.data.byuid) == null) {
            if (!window.hu60_user_info_map) {
                window.hu60_user_info_map = {}
            }
            if (!window.hu60_user_info_map[msg.data.byuid]) {
                fetch(window.hu60_site_url+'/q.php/user.info.'+msg.data.byuid+'.json?_origin=*')
                .then(res => res.json()).then(jres => {
                    window.hu60_user_info_map[msg.data.byuid] = jres
                    appendChatList({
                        uid: msg.data.byuid,
                        name: jres.name, 
                        avatar: window.hu60_site_file_url+"/avatar/"+msg.data.byuid+".jpg", 
                        isRobot: msg.data.byuid<0,
                        overview: msgContentText,
                        tipsCount: 0
                    }, {updateStorage: true})
                    appendChatText(msgContentText, msg.data.byuid, {self: false, updateStorage: true})
                })
                return
            }
            let userInfo = window.hu60_user_info_map[msg.data.byuid]
            appendChatList({
                uid: msg.data.byuid,
                name: userInfo.name, 
                avatar: window.hu60_site_file_url+"/avatar/"+msg.data.byuid+".jpg", 
                isRobot: msg.data.byuid<0,
                overview: msgContentText,
                tipsCount: 0
            }, {updateStorage: true})
        }
        appendChatText(msgContentText, msg.data.byuid, {self: false, updateStorage: true})
        return
    }

    console.debug('discard non chat message: ', event.data)
}


function getHu60MsgText(msgContent) {
  	const validUnits = ["text","imgzh","mdpre"]
    let text = msgContent.filter(unit => validUnits.includes(unit.type) ).map(unit => {
    	if(unit.type == "text") {
        	return setext(unit.value)
        }
      	if(unit.type == "mdpre") {
        	return unit.data
        }
        if(unit.type == "mdcode") {
            return unit.quote + unit.lang + unit.data + unit.quote
        }
      	if(unit.type == "imgzh") {
        	return '<img style="max-width: 90%;display: block" src="'+unit.src+'" alt="'+unit.alt+'" />'
        }
    }).join("").trim()
    if (text.startsWith(',') || text.startsWith('，')) {
        text = text.substr(1)
    }
    console.debug(text)
    return text
}


function onSmallScreenDevice() {
    let hu60botChat = document.querySelector('#hu60botChat')
    let chatList = document.querySelector('#chatList')
    let chatWindow = document.querySelector('#chatWindow')
    let chatContainer = document.querySelector('#chatContainer')
    let chatReturnBtn = document.querySelector('#chatReturnBtn')

    let winHeight = window.innerHeight


    hu60botChat.style.cssText = 'height: ' + winHeight + 'px'
    chatContainer.style.cssText = 'height: ' + (winHeight - 130) + 'px'
    chatList.style.cssText = 'display: none'

    chatReturnBtn.style.display = 'block'


    document.querySelectorAll('#chatList li').forEach(chatItem => chatItem.addEventListener('click', e=>{
        document.querySelector('#chatWindow').style.display = 'block'
        document.querySelector('#chatList').style.display = 'none'
    }))
}

function setext(unsestr) {
    return unsestr.replaceAll('<', '&lt;').replaceAll('>','&gt;').replaceAll('\n', '<br />').replaceAll(' ', '&nbsp;')
}

startPlugin()

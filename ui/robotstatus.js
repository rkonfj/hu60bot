(function() {
    let paramsSearch = new URLSearchParams(window.location.search)
    if(!paramsSearch.get('showBot') || paramsSearch.get('showBot') == 1) {
        let chatForm = document.querySelector('.chat-form')
        let commentsForm = document.querySelector('.comments-form')
        let topicForm = document.querySelector('.topic-form-content')

        let widgetForm = document.querySelector('.widget-form')
        let commentsReplay = document.querySelector('.comments-replay')
        let cr180Form = document.querySelector('.topic-form')

        let robotList = document.createElement('div')
        robotList.classList.add('robotList')
        robotList.innerHTML = `
        	<!-- author https://github.com/rkonfj -->
            <a href="javascript:;" onclick="tipsMe(this)" data-at="@ChatGPT" data-idx="0" title="@ChatGPT"><img src="https://file.hu60.cn/avatar/-50.jpg" class="avatar smallava" data-uid="-50" /></a>
            
            <a href="javascript:;" onclick="tipsMe(this)" data-at="@文心一言" data-idx="1" title="@文心一言"><img src="https://file.hu60.cn/avatar/-150.jpg" class="avatar smallava" data-uid="-150" /></a>
            
            <a href="javascript:;" onclick="tipsMe(this)" data-at="@通义千问" data-idx="2" title="@通义千问"><img src="https://img.alicdn.com/imgextra/i4/O1CN01c26iB51UyR3MKMFvk_!!6000000002586-2-tps-124-122.png" class="avatar smallava" data-uid="-200" /></a>
            
            <a href="javascript:;" onclick="tipsMe(this)" data-at="@讯飞星火" data-idx="3" title="@讯飞星火"><img src="https://file.hu60.cn/avatar/-300.jpg" class="avatar smallava" data-uid="-300" /></a>
            
            <a href="javascript:;" onclick="tipsMe(this)" data-at="@罐子2号" data-idx="4" title="@罐子2号"><img src="https://file.hu60.cn/avatar/-50.jpg" class="avatar smallava" data-uid="-51" /></a>
            
            <a href="javascript:;" onclick="tipsMe(this)" data-at="@靓仔" data-idx="5" title="@靓仔"><img src="https://file.hu60.cn/avatar/-50.jpg" class="avatar smallava" data-uid="-53" /></a>
            
            <a href="javascript:;" onclick="tipsMe(this)" data-at="@QA" data-idx="6" title="@QA"><img src="https://file.hu60.cn/avatar/-50.jpg" class="avatar smallava" data-uid="-55" /></a>
            
            <a href="javascript:;" onclick="tipsMe(this)" data-at="@Chatbot" data-idx="7" title="@Chatbot"><img src="https://file.hu60.cn/avatar/-56.jpg" class="avatar smallava" data-uid="-56" /></a>
            
            <a href="javascript:;" onclick="tipsMe(this)" data-at="@GPTbot" data-idx="8" title="@GPTbot"><img src="https://file.hu60.cn/avatar/-57.jpg" class="avatar smallava" data-uid="-57" /></a>
            
            <a href="javascript:;" onclick="tipsMe(this)" data-at="@yiyan" data-idx="9" title="@yiyan"><img src="https://file.hu60.cn/avatar/-150.jpg" class="avatar smallava" data-uid="-151" /></a>
        `
        if(chatForm) {
            widgetForm.insertBefore(robotList, chatForm)
        }

        if(commentsForm) {
            commentsReplay.insertBefore(robotList, commentsForm)
        }

        if(topicForm) {
            cr180Form.insertBefore(robotList, topicForm)
        }
    }

    let avatarList = document.querySelectorAll('.avatar')
    setInterval(()=>{
        avatarList.forEach(item => {
            if(!item.parentElement) {
                return
            }
            let u = new URL(item.src)
            let _uid = item.getAttribute('data-uid')
            let uid = _uid?_uid:u.pathname.match(/[-]?\d+/g)
            if(uid == null || uid > 0) {
                return
            }
            let color = `${window.hu60_hu60bot_online_user&&window.hu60_hu60bot_online_user[uid]&&window.hu60_hu60bot_online_user[uid]>0?'#1bbe36':'#ccc'}`
            let robotstatus = item.parentElement.querySelector('.robotstatus')
            if(robotstatus) {
                robotstatus.style.backgroundColor = color
                return
            }
            item.parentElement.style.position='relative'
            item.parentElement.style.display='inline-block'
            robotstatus = document.createElement('span') 
            robotstatus.classList.add('robotstatus')
            item.parentElement.appendChild(robotstatus)
            item.parentElement.querySelector('.robotstatus').style.backgroundColor = color
        })
    }, 300)

    document.addEventListener("keydown", function(event){
        if(event.ctrlKey) {
            window.hu60_robotstatus_multi_mode = true
            if(!window.hu60_robotstatus_multi_mode_buffer) {
                window.hu60_robotstatus_multi_mode_buffer = ""
            }
            console.debug("enter multi_mode")
            return
        }
        if(event.shiftKey) {
            window.hu60_robotstatus_shift_multi_mode = true
            console.debug("enter shift_multi_mode")
            return
        }
    })

    document.addEventListener("keyup", function(event){
        if(event.key == "Control") {
            window.hu60_robotstatus_multi_mode = false
            window.hu60_robotstatus_multi_mode_buffer = null
            console.debug("leave multi_mode")
            return
        }
        if(event.key == "Shift") {
            window.hu60_robotstatus_shift_multi_mode = false
            window.hu60_robotstatus_shift_multi_mode_start = null
            console.debug("leave shift_multi_mode")
            return
        }
    })
})()

function tipsMe(eventSource) {
    let at = eventSource.getAttribute('data-at')
    let idx = eventSource.getAttribute('data-idx')
    let robotsCur = `${at}，`

    if (window.hu60_robotstatus_multi_mode) {
        let thisAt = `${at}，`
        if(window.hu60_robotstatus_multi_mode_buffer.includes(thisAt)) {
            window.hu60_robotstatus_multi_mode_buffer = window.hu60_robotstatus_multi_mode_buffer.replace(thisAt, "")
        } else {
            window.hu60_robotstatus_multi_mode_buffer += thisAt
        }
        tipsRobots(window.hu60_robotstatus_multi_mode_buffer)
        return
    }

    if(window.hu60_robotstatus_shift_multi_mode) {
        if(!window.hu60_robotstatus_shift_multi_mode_start) {
            window.hu60_robotstatus_shift_multi_mode_start = idx
            tipsRobots(robotsCur)
            return
        }
        if(window.hu60_robotstatus_shift_multi_mode_start == idx) {
            tipsRobots("")
            return
        }
        window.hu60_robotstatus_multi_mode_buffer = getRobots(window.hu60_robotstatus_shift_multi_mode_start, idx)
        tipsRobots(window.hu60_robotstatus_multi_mode_buffer)
        return
    }

    if(window.hu60_robotstatus_multi_mode_last_robots == robotsCur) {
        tipsRobots("")
        window.hu60_robotstatus_shift_multi_mode_start = null
        if(window.hu60_robotstatus_multi_mode_buffer) {
            window.hu60_robotstatus_multi_mode_buffer = window.hu60_robotstatus_multi_mode_buffer.replace(robotsCur, "")
        }
    } else {
        tipsRobots(robotsCur)
        window.hu60_robotstatus_shift_multi_mode_start = idx
        window.hu60_robotstatus_multi_mode_buffer = robotsCur
    }
}

function tipsRobots(robotsCur) {
    let robotsLast = window.hu60_robotstatus_multi_mode_last_robots
    let inputContent = document.querySelector('#content')
    if(inputContent.value.startsWith(robotsLast)) {
        inputContent.value = inputContent.value.substring(robotsLast.length)
    }
    inputContent.value = `${robotsCur}${inputContent.value}`
    window.hu60_robotstatus_multi_mode_last_robots = robotsCur
    inputContent.focus()
}


function getRobots(start, stop) {
    if(start > stop) {
        [ start, stop ] = [ stop, start ]
    }
    let robots = ""
    document.querySelectorAll(".robotList a").forEach(robot => {
        let idx =robot.getAttribute("data-idx")
        if(idx >= start && idx <= stop) {
            let at =robot.getAttribute("data-at")
            robots += `${at}，`
        }
    })
    return robots
}
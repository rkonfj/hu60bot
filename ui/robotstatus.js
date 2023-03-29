 (function() {
    let chatForm = document.querySelector('.chat-form')
    let commentsForm = document.querySelector('.comments-form')
    let topicForm = document.querySelector('.topic-form-content')

    let widgetForm = document.querySelector('.widget-form')
    let commentsReplay = document.querySelector('.comments-replay')
    let cr180Form = document.querySelector('.topic-form')

    let robotList = document.createElement('div')
    robotList.classList.add('robotList')
    robotList.innerHTML = `
        <a href="javascript:;" onclick="tipsMe(this)" data-at="@hu60bot"><img src="https://file.hu60.cn/avatar/-50.jpg" class="avatar smallava" data-uid="-54" /></a>
        <a class="txtlink" href="javascript:;" onclick="tipsMe(this)" data-at="@hu60bot">@hu60bot</a>
        <a href="javascript:;" onclick="tipsMe(this)" data-at="@ChatGPT"><img src="https://file.hu60.cn/avatar/-50.jpg" class="avatar smallava" data-uid="-50" /></a>
        <a class="txtlink" onclick="tipsMe(this)" href="javascript:;" data-at="@ChatGPT">@ChatGPT</a>
        <a href="javascript:;" onclick="tipsMe(this)" data-at="@罐子2号"><img src="https://file.hu60.cn/avatar/-50.jpg" class="avatar smallava" data-uid="-51" /></a>
        <a class="txtlink" onclick="tipsMe(this)" href="javascript:;" data-at="@罐子2号">@罐子2号</a>
        <a href="javascript:;" onclick="tipsMe(this)" data-at="@靓仔"><img src="https://file.hu60.cn/avatar/-50.jpg" class="avatar smallava" data-uid="-53" /></a>
        <a class="txtlink" onclick="tipsMe(this)" href="javascript:;" data-at="@靓仔">@靓仔</a>
        <a href="javascript:;" onclick="tipsMe(this)" data-at="@QA"><img src="https://file.hu60.cn/avatar/-50.jpg" class="avatar smallava" data-uid="-55" /></a>
        <a class="txtlink" onclick="tipsMe(this)" href="javascript:;" data-at="@QA">@QA</a>
        <a href="javascript:;" onclick="tipsMe(this)" data-at="@Chatbot"><img src="https://file.hu60.cn/avatar/-56.jpg" class="avatar smallava" data-uid="-56" /></a>
        <a class="txtlink" onclick="tipsMe(this)" href="javascript:;" data-at="@Chatbot">@Chatbot</a>
        <a href="javascript:;" onclick="tipsMe(this)" data-at="@GPTbot"><img src="https://file.hu60.cn/avatar/-57.jpg" class="avatar smallava" data-uid="-57" /></a>
        <a class="txtlink" onclick="tipsMe(this)" href="javascript:;" data-at="@GPTbot">@GPTbot</a>
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

})()

function tipsMe(eventSource) {
    let at = eventSource.getAttribute('data-at')
    if(document.querySelector('#content').value.startsWith(at)) {
        window.hu60_robotstatus_at = null
        document.querySelector('#content').value = document.querySelector('#content').value.substring(at.length+1)
        document.querySelector('#content').focus()
        return
    }
    if(window.hu60_robotstatus_at && document.querySelector('#content').value.startsWith(window.hu60_robotstatus_at)) {
        document.querySelector('#content').value = document.querySelector('#content').value.substring(window.hu60_robotstatus_at.length)
        document.querySelector('#content').value = `${at}${document.querySelector('#content').value}`
        window.hu60_robotstatus_at = at
        document.querySelector('#content').focus()
        return
    }
    window.hu60_robotstatus_at = at
    document.querySelector('#content').value = `${at}，${document.querySelector('#content').value}`
    document.querySelector('#content').focus()
}
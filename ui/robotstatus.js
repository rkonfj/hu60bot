 (function() {
    let avatarList = document.querySelectorAll('.avatar')
    setInterval(()=>{
        avatarList.forEach(item => {
            if(!item.parentElement) {
                return
            }
            let u = new URL(item.src)
            let uid = u.pathname.match(/[-]?\d+/g)
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
    }, 200)
})()

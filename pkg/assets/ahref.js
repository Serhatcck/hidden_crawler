(() => {
    const urls = [];
    document.querySelectorAll('a[href]').forEach(el => {
        if (el.href) urls.push(el.href);
        if (el.src) urls.push(el.src);
    });
    return urls;
})()
function f() {
    let data = {"name": $("#name").val()}
    $.ajax({
        url: 'http://localhost:8080/user/signup',
        type: 'GET',
        data: JSON.stringify(data),
        contentType: "application/javascript; charset=utf-8",
        dataType: "json",
    });
    $("#text1").append(data)
}
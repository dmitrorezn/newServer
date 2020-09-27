function sendData() {
    let data = {
        "name": $('#name').val(),
    }
    $.ajax({
        url: '/data/show',
        type: 'POST',
        data: JSON.stringify(data),
        contentType: "application/json; charset=utf-8",
        dataType: "json",
        success : function(data) {
            $('#brd').append("<div>"+ data.name +"</div>");
        },
    });
    console.log(data)
}
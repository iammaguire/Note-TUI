<?php
    include 'db.php';

    if(isset($_GET['title']) && isset($_GET['content']) && isset($_GET['id'])) {
        $sql = "update notebook set title = '" . $_GET['title'] . "', content = '" . $_GET['content'] . "' where id = " . $_GET['id'] . "";
        $conn = new mysqli($server, $username, $password, $database);

        if ($conn->connect_errno) {
            print "DB Connection Error.";
            http_response_code(404);
            die();
        }

        if($conn->query($sql)) {
            print "Success.";
            http_response_code(200);
        } else {
            print "Query failed.";
            http_response_code(404);
        }

        $conn->close();
    } else {
        print "Title, content or id not set.";
        http_response_code(404);
    }
?>
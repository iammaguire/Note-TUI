<?php
    include 'db.php';

    $conn = new mysqli($server, $username, $password, $database);
    
    if ($conn->connect_errno) {
        http_response_code(404);
        die();
    }

    $data = $conn->query('select * from notebook');

    if($data) {
        $arr = array();
        while($row = mysqli_fetch_assoc($data)) {
            $arr[] = $row;
        }
    
        print json_encode($arr);
        http_response_code(200);
    } else {
        http_response_code(404);
    }

    $conn->close();
?>

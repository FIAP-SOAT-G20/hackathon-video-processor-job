Feature: Video processing via Lambda
  As a client of the video processor
  I want to invoke the Lambda with a video key
  So that I get a processing response

  Scenario: Missing video_key returns 400
    Given a lambda event with video_key ""
    When I invoke the lambda handler
    Then the response statusCode is 400
    And the response JSON has field "success" equal to false

  Scenario: Successful processing returns 200
    Given a lambda event with video_key "sample/video.mp4"
    And the controller returns success with frame_count 12 and output_key "processed/sample/video_frames.zip"
    When I invoke the lambda handler
    Then the response statusCode is 200
    And the response JSON has field "success" equal to true
    And the response JSON has field "frame_count" equal to 12
    And the response JSON has field "output_key" equal to "processed/sample/video_frames.zip"

  Scenario: Successful processing with hash returns 200
    Given a lambda event with video_key "sample/video.mp4"
    And the controller returns success with frame_count 12 and output_key "processed/e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855.zip" and hash "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
    When I invoke the lambda handler
    Then the response statusCode is 200
    And the response JSON has field "success" equal to true
    And the response JSON has field "frame_count" equal to 12
    And the response JSON has field "output_key" equal to "processed/e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855.zip"
    And the response JSON has field "hash" equal to "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

  Scenario: Processing with custom configuration sanitization
    Given a lambda event with video_key "sample/video.mp4" and configuration frame_rate 0 and output_format "JPG"
    And the controller returns success with frame_rate sanitized to 1.0 and format "jpg" and frame_count 5
    When I invoke the lambda handler
    Then the response statusCode is 200
    And the response JSON has field "success" equal to true
    And the response JSON has field "frame_count" equal to 5
    And the response JSON has field "output_key" contains "processed/"
    And the response JSON has field "hash" is not empty

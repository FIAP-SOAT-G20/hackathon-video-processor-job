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

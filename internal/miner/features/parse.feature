Feature: The parser validation
    In order to work with responses from the pool
    As a noso pool client
    I need to make sure the pool responses are parsed correctly

    Background:
        Given I have a new comms object
          And I have a pool IP of "1.2.3.4"
          And I have a wallet address of "012345678901234567890123456789"
          And I the current block is 1234

    Scenario: JOINOK_01 parsed correctly
         When I parse the "JOINOK_01" response
         Then the comms.PoolAddr channel should have "JOINOK_01POOLADDRESS"
          And the comms.MinerSeed channel should have "!3!!!!!!!"
          And the comms.Block channel should have 5891
          And the comms.TargetString channel should have "JOINOK_01TARGETSTRING"
          And the comms.TargetChars channel should have 11
          And the comms.Step channel should have 0
          And the comms.Diff channel should have 102
          And the comms.Balance channel should have "0"
          And the comms.BlocksTillPayment channel should have "-4"
          And the comms.Joined channel should be called

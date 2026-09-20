module test
(
    input clk,
    input i_rst,
    input misc,
    input ternarymisc,
    input i_pulse,
    output logic o_misc,
    output logic t_misc
);

logic i_flag = 1;
logic dummy_var;



always_ff @(posedge clk)
begin
    if(i_rst)
    begin
        dummy_var <= 0;
    end
    else
    begin

        if(!i_pulse)
        begin

            if(!i_flag)
            begin
                i_flag <= 1;
            end
            
        end

    end

    
        
    
end

always_comb
begin
    assign o_misc = misc;
    assign t_misc = ternarymisc ? 1'b1 : 1'b0;
end

assert property (@(posedge clk) !i_pulse |=> ##3 i_flag == 1);


property test_rst_constraint;
    @(posedge clk) !i_rst;
endproperty


assume property (test_rst_constraint);

endmodule